import hashlib
from pathlib import Path
from typing import Any, Dict
from urllib.parse import urlparse

import numpy as np
import requests
from google.cloud import storage


def sha256_file(path: Path) -> str:
    h = hashlib.sha256()
    with path.open("rb") as f:
        for chunk in iter(lambda: f.read(8192), b""):
            h.update(chunk)
    return h.hexdigest()


def download_file(url: str, type: str, dest: Path, expected_sha256: str = ""):
    dest.parent.mkdir(parents=True, exist_ok=True)

    if dest.exists() and expected_sha256 and sha256_file(dest) == expected_sha256:
        return dest
    match type:
        case "gcs":
            p = urlparse(url)
            bucket_name = p.netloc
            blob_name = p.path.lstrip("/")

            client = storage.Client()
            bucket = client.bucket(bucket_name)
            blob = bucket.blob(blob_name)

            if blob.exists():
                blob.download_to_filename(dest)

    if expected_sha256 and sha256_file(dest) != expected_sha256:
        raise RuntimeError(f"SHA256 mismatch for {dest}")

    return dest


def get_manifest(registry: str, model_name: str, model_version: str):
    url = f"{registry}/v1/models/{model_name}?version={model_version}"
    res = requests.get(url)
    res.raise_for_status()

    return res.json()


def get_artifacts(manifest, cache_dir: Path):
    artifacts = manifest.get("artifacts", {})
    model_dir = (
        cache_dir / f"{manifest['model']['name']}-{manifest['model']['version']}"
    )
    local_paths = {}

    for key, info in artifacts.items():
        url = info["url"]
        sha = info["sha256"]
        type = info["type"]
        filename = Path(url).name
        dest = model_dir / filename
        download_file(url, type, dest, expected_sha256=sha)
        local_paths[key] = dest

    return local_paths


def load(*, registry: str, model_name: str, model_version: str, cache_dir: Path):
    manifest = get_manifest(
        registry=registry, model_name=model_name, model_version=model_version
    )

    files = get_artifacts(manifest=manifest, cache_dir=cache_dir)

    return files, manifest


class BaseModelLoader:
    def load(self, model_dir: Path) -> None:
        raise NotImplementedError()

    def infer(self, waveform: np.ndarray, sample_rate: int) -> Dict[str, Any]:
        raise NotImplementedError()
