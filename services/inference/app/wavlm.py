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
        self.model = WavLMForSequenceClassification.from_pretrained(model_dir)
        self.feature_extractor = AutoFeatureExtractor.from_pretrained(model_dir)

        assert self.model is not None
        assert self.feature_extractor is not None

        self.model.to(self.device)
        self.model.eval()

    def infer(self, waveform: np.ndarray, sample_rate: int):
        assert self.model is not None
        assert self.feature_extractor is not None

        inputs = self.feature_extractor(
            waveform,
            sampling_rate=sample_rate,
            return_tensors="pt",
            padding=True,
        )

        inputs = {k: v.to(self.device) for k, v in inputs.items()}

        with torch.no_grad():
            outputs = self.model(**inputs)
            logits = outputs.logits

        logits = logits.detach().cpu().numpy()[0]
        pred = int(np.argmax(logits))

        return {
            "prediction": pred,
            "scores": logits.tolist(),
        }
