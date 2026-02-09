import hashlib
from pathlib import Path

import requests


def sha256_file(path: Path) -> str:
    h = hashlib.sha256()
    with path.open("rb") as f:
        for chunk in iter(lambda: f.read(8192), b""):
            h.update(chunk)
    return h.hexdigest()


def download_file(url: str, dest: Path, expected_sha256: str = ""):
    dest.parent.mkdir(parents=True, exist_ok=True)

    if dest.exists() and expected_sha256 and sha256_file(dest) == expected_sha256:
        return dest

    resp = requests.get(url, stream=True)
    resp.raise_for_status()
    with open(dest, "wb") as f:
        for chunk in resp.iter_content(8192):
            f.write(chunk)

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
    model_dir = cache_dir / manifest["model"]["name"] / manifest["model"]["version"]
    local_paths = {}

    for key, info in artifacts.items():
        url = info["url"]
        sha = info.get("sha256") or manifest["model"].get("sha256")
        filename = Path(url).name
        dest = model_dir / filename
        download_file(url, dest, expected_sha256=sha)
        local_paths[key] = dest

    return local_paths


def load(*, registry: str, model_name: str, model_version: str, cache_dir: Path):
    manifest = get_manifest(
        registry=registry, model_name=model_name, model_version=model_version
    )

    files = get_artifacts(manifest=manifest, cache_dir=cache_dir)

    return files, manifest


class BaseModelLoader:
    def load(self, model_dir: Path):
        raise NotImplementedError()

    def infer(self, inputs: dict) -> dict:
        raise NotImplementedError()
