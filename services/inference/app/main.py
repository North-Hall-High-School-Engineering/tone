import os
from pathlib import Path

from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from transformers import AutoFeatureExtractor, WavLMForSequenceClassification

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
model = WavLMForSequenceClassification.from_pretrained(str(model_dir))
feature_extractor = AutoFeatureExtractor.from_pretrained(str(model_dir))

loader = get_model_loader(MODEL_NAME, MODEL_VERSION)
loader.load(model_dir)


class Request(BaseModel):
    input_values: list
    attention_mask: list


class Response(BaseModel):
    prediction: int
    scores: list


@app.post("/v1/infer", response_model=Response)
async def infer(req: Request):
    try:
        result = loader.infer(
            {"input_values": req.input_values, "attention_mask": req.attention_mask}
        )
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

    return Response(prediction=result["prediction"], scores=result["scores"])
