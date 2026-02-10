import os
from pathlib import Path

import numpy as np
import onnxruntime as ort
import torch
from transformers import AutoFeatureExtractor, WavLMForSequenceClassification

from .model import BaseModelLoader


class WavLMTorchLoader(BaseModelLoader):
    def __init__(self, device: str = "cpu"):
        self.device = torch.device(device)
        self.model: WavLMForSequenceClassification | None = None
        self.feature_extractor: AutoFeatureExtractor | None = None

    def load(self, model_dir: Path) -> None:
        self.model = WavLMForSequenceClassification.from_pretrained(model_dir)
        self.feature_extractor = AutoFeatureExtractor.from_pretrained(model_dir)

        self.model.to(self.device)
        self.model.eval()

    def infer(self, waveform: np.ndarray, sample_rate: int) -> dict:
        if self.model is None or self.feature_extractor is None:
            raise RuntimeError("Model not loaded")

        inputs = self.feature_extractor(
            waveform,
            sampling_rate=sample_rate,
            return_tensors="pt",
        )

        inputs = {k: v.to(self.device) for k, v in inputs.items()}

        with torch.inference_mode():
            logits = self.model(**inputs).logits

        logits_np = logits.squeeze(0).cpu().numpy()
        pred = int(np.argmax(logits_np))

        return {
            "prediction": pred,
            "scores": logits_np.tolist(),
        }


class WavLMOnnxLoader(BaseModelLoader):
    def __init__(self):
        self.session: ort.InferenceSession | None = None
        self.feature_extractor: AutoFeatureExtractor | None = None

    def load(self, model_dir: Path) -> None:
        sess_options = ort.SessionOptions()
        sess_options.graph_optimization_level = (
            ort.GraphOptimizationLevel.ORT_ENABLE_ALL
        )
        sess_options.intra_op_num_threads = os.cpu_count() or 1
        sess_options.inter_op_num_threads = 2

        self.session = ort.InferenceSession(
            str(model_dir / "model.onnx"),
            sess_options=sess_options,
            providers=["CPUExecutionProvider"],
        )

        self.feature_extractor = AutoFeatureExtractor.from_pretrained(model_dir)

        dummy_waveform = np.zeros((16000,), dtype=np.float32)
        dummy_inputs = self.feature_extractor(
            dummy_waveform, sampling_rate=16000, return_tensors="np"
        )

        ort_inputs = {}
        for k, v in dummy_inputs.items():
            if np.issubdtype(v.dtype, np.integer):
                ort_inputs[k] = v.astype(np.int64)
            else:
                ort_inputs[k] = v

        _ = self.session.run(None, ort_inputs)

    def infer(self, waveform: np.ndarray, sample_rate: int) -> dict:
        if self.session is None or self.feature_extractor is None:
            raise RuntimeError("Model not loaded")

        inputs = self.feature_extractor(
            waveform,
            sampling_rate=sample_rate,
            return_tensors="np",
        )

        ort_inputs = {}
        for k, v in inputs.items():
            if np.issubdtype(v.dtype, np.integer):
                ort_inputs[k] = v.astype(np.int64)
            else:
                ort_inputs[k] = v

        logits = self.session.run(None, ort_inputs)[0][0]

        pred = int(np.argmax(logits))

        return {
            "prediction": pred,
            "scores": logits.tolist(),
        }
