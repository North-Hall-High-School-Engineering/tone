from pathlib import Path
from typing import Any, Dict

import numpy as np


class BaseModelLoader:
    def load(self, model_dir: Path) -> None:
        raise NotImplementedError()

    def infer(self, waveform: np.ndarray, sample_rate: int) -> Dict[str, Any]:
        raise NotImplementedError()
