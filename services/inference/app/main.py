import asyncio
import io
import os
import uuid
from collections import deque
from pathlib import Path

import numpy as np
import soundfile as sf
import torch
from fastapi import (
    FastAPI,
    File,
    HTTPException,
    UploadFile,
    WebSocket,
    WebSocketDisconnect,
)
from scipy.signal import resample_poly
from silero_vad import get_speech_timestamps, load_silero_vad, read_audio

from .loader import load
from .registry import get_model_loader

app = FastAPI()

REGISTRY_URL = os.getenv("REGISTRY_URL", "http://tone-registry-service:80")
MODEL_NAME = os.getenv("MODEL_NAME", "tone")
MODEL_VERSION = os.getenv("MODEL_VERSION", "1.0.1")
CACHE_DIR = Path(os.getenv("MODEL_CACHE_DIR", "/cache/models"))

artifacts, manifest = load(
    REGISTRY_URL,
    MODEL_NAME,
    MODEL_VERSION,
    CACHE_DIR,
)
model_dir = Path(artifacts["model"]).parent
loader = get_model_loader(manifest)
loader.load(model_dir)

silero, utils = torch.hub.load(
    repo_or_dir="snakers4/silero-vad",
    model="silero_vad",
    force_reload=False,
    trust_repo=True,
)
(get_speech_timestamps, save_audio, read_audio, VADIterator, collect_chunks) = utils

SAMPLE_RATE = 16000
CHUNK_SIZE = 512
PRE_ROLL_MS = 100
FRAME_MS = (CHUNK_SIZE / SAMPLE_RATE) * 1000
NUM_PRE_ROLL_FRAMES = int(PRE_ROLL_MS // FRAME_MS)
SILENCE_CHUNKS_TO_END = int(0.3 * SAMPLE_RATE / CHUNK_SIZE)


@app.websocket("/v1/ws")
async def stream(websocket: WebSocket):
    await websocket.accept()

    triggered = False

    ring_buffer = deque(maxlen=NUM_PRE_ROLL_FRAMES)
    audio_buffer = bytearray()
    vad_buffer = np.array([], dtype=np.float32)

    vad_iterator = VADIterator(silero)

    try:
        while True:
            data = await websocket.receive_bytes()
            if not data:
                break

            pcm_samples = np.frombuffer(data, dtype=np.float32)
            if pcm_samples.size == 0:
                continue
            vad_buffer = np.concatenate((vad_buffer, pcm_samples))

            while len(vad_buffer) >= CHUNK_SIZE:
                chunk = vad_buffer[:CHUNK_SIZE]
                vad_buffer = vad_buffer[CHUNK_SIZE:]

                speech_events = vad_iterator(chunk, return_seconds=False)
                is_speech_start = speech_events is not None and "start" in speech_events
                is_speech_end = speech_events is not None and "end" in speech_events

                chunk_bytes = (chunk * 32768.0).astype(np.int16).tobytes()

                if is_speech_start and not triggered:
                    for rb_chunk in ring_buffer:
                        audio_buffer.extend(rb_chunk)
                    ring_buffer.clear()
                    triggered = True

                if triggered:
                    audio_buffer.extend(chunk_bytes)

                    if is_speech_end:
                        utterance_np = (
                            np.frombuffer(audio_buffer, dtype=np.int16).astype(
                                np.float32
                            )
                            / 32768.0
                        )
                        results = loader.infer(utterance_np, 16000)
                        await websocket.send_json(results)
                        audio_buffer.clear()
                        triggered = False
                else:
                    ring_buffer.append(chunk_bytes)

    except WebSocketDisconnect:
        print("WebSocket disconnected")
    finally:
        await websocket.close()
