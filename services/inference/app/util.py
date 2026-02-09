from .loader import BaseModelLoader
from .wavlm import WavLMModelLoader


def get_model_loader(model_name: str, version: str) -> BaseModelLoader:
    if model_name == "tone":
        match version:
            case "1.0.0":
                return WavLMModelLoader()
            case _:
                return WavLMModelLoader()

    else:
        raise ValueError(f"No loader implemented for {model_name} v{version}")
