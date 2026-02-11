import io
import os
from pathlib import Path

import numpy as np
import soundfile as sf
import webrtcvad
from fastapi import (
    FastAPI,
    File,
    HTTPException,
    UploadFile,
    WebSocket,
    WebSocketDisconnect,
)
from scipy.signal import resample_poly

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


@app.post("/v1/infer")
async def infer(file: UploadFile = File(...)):
    try:
        audio_bytes = await file.read()
        waveform, sr = sf.read(io.BytesIO(audio_bytes))

        if waveform.ndim > 1:
            waveform = waveform.mean(axis=1)

        if sr != 16000:
            gcd = np.gcd(sr, 16000)
            up = 16000 // gcd
            down = sr // gcd
            waveform = resample_poly(waveform, up, down)
            sr = 16000

        waveform = waveform.astype(np.float32)
        result = loader.infer(waveform, sr)

        return {"prediction": result["prediction"], "scores": result["scores"]}

    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))
