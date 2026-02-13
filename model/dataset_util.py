import glob
import os
import shutil
import urllib.request
import zipfile

from datasets import Audio, Dataset
from tqdm import tqdm

RAVDESS_EMOTIONS = {
    "01": "neutral",
    # "02": "calm",
    "03": "happy",
    "04": "sad",
    "05": "angry",
    "06": "fearful",
    "07": "disgust",
    # "08": "surprised",
}

TESS_EMOTIONS = {
    "neutral": "neutral",
    # "calm": "calm",
    "happy": "happy",
    "sad": "sad",
    "angry": "angry",
    "fear": "fearful",
    "disgust": "disgust",
    # "ps": "surprised",
}

CREMA_EMOTIONS = {
    "NEU": "neutral",
    "HAP": "happy",
    "SAD": "sad",
    "ANG": "angry",
    "FEA": "fearful",
    "DIS": "disgust",
}

EMODB_EMOTIONS = {
    "W": "angry",
    # "L": "boredom",
    "E": "disgust",
    "A": "fearful",
    "F": "happy",
    "T": "sad",
    "N": "neutral",
}


def download_ravdess(base_dir="data"):
    return download_and_extract(
        url="https://zenodo.org/record/1188976/files/Audio_Speech_Actors_01-24.zip",
        archive_path=os.path.join(base_dir, "ravdess.zip"),
        extract_dir=os.path.join(base_dir, "ravdess"),
        dataset_name="RAVDESS",
    )


def download_tess(base_dir="data"):
    return download_and_extract(
        url="https://www.kaggle.com/api/v1/datasets/download/ejlok1/toronto-emotional-speech-set-tess",
        archive_path=os.path.join(base_dir, "tess.zip"),
        extract_dir=os.path.join(base_dir, "tess"),
        dataset_name="TESS",
    )


def download_cremad(base_dir="data"):
    return download_and_extract(
        url="https://www.kaggle.com/api/v1/datasets/download/ejlok1/cremad",
        archive_path=os.path.join(base_dir, "crema-d.zip"),
        extract_dir=os.path.join(base_dir, "CREMA-D"),
        dataset_name="CREMA-D",
    )


def download_emodb(base_dir="data"):
    return download_and_extract(
        url="https://www.kaggle.com/api/v1/datasets/download/piyushagni5/berlin-database-of-emotional-speech-emodb",
        archive_path=os.path.join(base_dir, "emodb.zip"),
        extract_dir=os.path.join(base_dir, "emodb"),
        dataset_name="EMODB",
    )


def cleanup(paths):
    for path in paths:
        try:
            if os.path.isdir(path):
                shutil.rmtree(path)
            elif os.path.isfile(path):
                os.remove(path)
        except Exception as e:
            print(f"Cleanup failed for {path}: {e}")


def download_and_extract(url, archive_path, extract_dir, dataset_name="dataset"):
    os.makedirs(os.path.dirname(archive_path), exist_ok=True)
    os.makedirs(extract_dir, exist_ok=True)

    if os.path.isdir(extract_dir) and not len(os.listdir(extract_dir)) == 0:
        print(f"{dataset_name} directory already exists. Skipping download.")
        return extract_dir

    try:
        print(f"{dataset_name} not found. Downloading...")

        # Download with progress bar
        with tqdm(
            unit="B", unit_scale=True, miniters=1, desc=f"Downloading {dataset_name}"
        ) as t:
            last = 0

            def reporthook(count, block_size, total_size):
                nonlocal last
                if t.total is None:
                    t.total = total_size
                downloaded = count * block_size
                delta = downloaded - last
                last = downloaded
                if delta > 0:
                    t.update(delta)

            urllib.request.urlretrieve(url, archive_path, reporthook)

        # Extract
        with zipfile.ZipFile(archive_path, "r") as archive:
            members = archive.namelist()
            for member in tqdm(members, desc=f"Extracting {dataset_name}"):
                archive.extract(member, extract_dir)

        print(f"{dataset_name} downloaded and extracted successfully.")
        return extract_dir

    except Exception as e:
        print(f"Error downloading or extracting {dataset_name}: {e}")
        cleanup([extract_dir])
        return None

    finally:
        if os.path.isfile(archive_path):
            try:
                os.remove(archive_path)
            except Exception as e:
                print(f"Failed to remove archive {archive_path}: {e}")


def load_ravdess_dataset(root):
    samples = []
    for wav in glob.glob(os.path.join(root, "*.wav"), recursive=True):
        fname = os.path.basename(wav)
        parts = fname.split("-")
        if len(parts) != 7:
            continue
        emotion = RAVDESS_EMOTIONS.get(parts[2])
        if emotion:
            samples.append({"audio": wav, "emotion": emotion})
    return samples


def load_tess_dataset(root):
    samples = []
    for wav in glob.glob(os.path.join(root, "**", "*.wav"), recursive=True):
        fname = os.path.basename(wav).lower()
        for key, emo in TESS_EMOTIONS.items():
            if key in fname:
                samples.append({"audio": wav, "emotion": emo})
                break
    return samples


def load_crema_dataset(root):
    samples = []
    for wav in glob.glob(os.path.join(root, "**", "*.wav")):
        parts = os.path.basename(wav).split("_")
        if len(parts) < 3:
            continue
        emotion = CREMA_EMOTIONS.get(parts[2])
        if emotion:
            samples.append({"audio": wav, "emotion": emotion})
    return samples


def load_emodb_dataset(root):
    samples = []
    for wav in glob.glob(os.path.join(root, "**", "*.wav")):
        emotion_code = os.path.basename(wav)[5]
        emotion = EMODB_EMOTIONS.get(emotion_code)
        if emotion:
            samples.append({"audio": wav, "emotion": emotion})
    return samples


def load_all_datasets(
    ravdess_dir=None,
    tess_dir=None,
    cremad_dir=None,
    emodb_dir=None,
):
    samples = []

    if ravdess_dir:
        samples += load_ravdess_dataset(ravdess_dir)
    if tess_dir:
        samples += load_tess_dataset(tess_dir)
    if cremad_dir:
        samples += load_crema_dataset(cremad_dir)
    if emodb_dir:
        samples += load_emodb_dataset(emodb_dir)

    # Create unified label mapping
    unique_emotions = sorted({s["emotion"] for s in samples})
    emotion2id = {e: i for i, e in enumerate(unique_emotions)}

    dataset_dict = {
        "audio": [s["audio"] for s in samples],
        "label": [emotion2id[s["emotion"]] for s in samples],
    }

    dataset = Dataset.from_dict(dataset_dict)
    dataset = dataset.cast_column("audio", Audio(sampling_rate=16000))

    return dataset, emotion2id
