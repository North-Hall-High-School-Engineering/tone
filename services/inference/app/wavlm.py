from pathlib import Path

import numpy as np
import torch
from transformers import AutoFeatureExtractor, WavLMForSequenceClassification

from .loader import BaseModelLoader


class WavLMModelLoader(BaseModelLoader):
    def __init__(self, device: str = "cpu"):
        self.model: WavLMForSequenceClassification | None = None
        self.feature_extractor: AutoFeatureExtractor | None = None
        self.device = torch.device(device)

    def load(self, model_dir: Path):
        self.model = WavLMForSequenceClassification.from_pretrained(
            model_dir, device_map=None
        )
        self.feature_extractor = AutoFeatureExtractor.from_pretrained(model_dir)

        assert self.model is not None
        assert self.feature_extractor is not None

        self.model.to(self.device)
        self.model.eval()

    def infer(self, inputs: dict):
        assert self.model is not None
        assert self.feature_extractor is not None

        input_values = torch.tensor([inputs["input_values"]], dtype=torch.float32).to(
            self.device
        )
        attention_mask = torch.tensor([inputs["attention_mask"]], dtype=torch.long).to(
            self.device
        )

        with torch.no_grad():
            logits = self.model(
                input_values=input_values, attention_mask=attention_mask
            ).logits
        logits = logits.cpu().numpy()[0]
        pred = int(np.argmax(logits))
        return {"prediction": pred, "scores": logits.tolist()}
