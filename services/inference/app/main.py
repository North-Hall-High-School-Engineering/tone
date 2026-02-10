import io
import os
from pathlib import Path

import torch
import torchaudio
from fastapi import FastAPI, File, HTTPException, UploadFile
from pydantic import BaseModel

from .loader import load
from .util import get_model_loader

app = FastAPI(title="tone inference service")

REGISTRY_URL = os.getenv("REGISTRY_URL", "http://registry:8080")
MODEL_NAME = os.getenv("MODEL_NAME", "tone")
MODEL_VERSION = os.getenv("MODEL_VERSION", "1.0.0")
CACHE_DIR = Path(os.getenv("MODEL_CACHE_DIR", "./models"))

artifacts, manifest = load(
    registry=REGISTRY_URL,
    model_name=MODEL_NAME,
    model_version=MODEL_VERSION,
    cache_dir=CACHE_DIR,
)

model_dir = Path(artifacts["model"]).parent

loader = get_model_loader(MODEL_NAME, MODEL_VERSION)
loader.load(model_dir)


class Request(BaseModel):
    input_values: list
    attention_mask: list


class Response(BaseModel):
    prediction: int
    scores: list


@app.post("/v1/infer", response_model=Response)
async def infer(file: UploadFile = File(...)):
    try:
        audio_bytes = await file.read()

        waveform, sr = torchaudio.load(io.BytesIO(audio_bytes))

        if sr != 16000:
            waveform = torchaudio.functional.resample(waveform, sr, 16000)

        waveform = waveform.mean(dim=0)
        waveform = waveform.numpy()
        result = loader.infer(waveform, 16000)

    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

    return Response(prediction=result["prediction"], scores=result["scores"])
