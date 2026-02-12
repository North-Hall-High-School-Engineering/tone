from .model import BaseModelLoader
from .wavlm import WavLMOnnxLoader, WavLMTorchLoader

MODEL_REGISTRY: dict[str, dict[str, dict[str, type[BaseModelLoader]]]] = {
    "tone": {
        "1.0.0": {
            "safetensors": WavLMTorchLoader,
            "onnx": WavLMOnnxLoader,
        },
        "1.0.1": {
            "onnx": WavLMOnnxLoader,
        },
        "1.0.2": {
            "onnx": WavLMOnnxLoader,
        },
    }
}


def get_model_loader(
    manifest,
) -> BaseModelLoader:
    name = manifest["model"]["name"]
    version = manifest["model"]["version"]
    format = manifest["model"]["format"]
    try:
        loader_cls = MODEL_REGISTRY[name][version][format]
    except KeyError:
        raise ValueError(f"No loader for {name} v{version} ({format})")

    return loader_cls()
