import glob
import os
import shutil
import urllib.request
import zipfile
import torch
import evaluate
import numpy as np 

from datasets import Audio, Dataset
from tqdm import tqdm

# !wget -q https://zenodo.org/record/1188976/files/Audio_Speech_Actors_01-24.zip
# !unzip -q Audio_Speech_Actors_01-24.zip -d ravdess

SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
accuracy = evaluate.load("accuracy")

def cleanup(paths):
    for path in paths:
        try:
            if os.path.isdir(path):
                shutil.rmtree(path)
            elif os.path.isfile(path):
                os.remove(path)
        except Exception as e:
            print(f"Cleanup failed for {path}: {e}")


def load_ravdess_dataset(ravdess_archive, ravdess_dir):
    if os.path.isdir(ravdess_dir):
        print("RAVDESS dataset directory already exists. Skipping download.")
        return ravdess_dir

    try:
        print("RAVDESS dataset not found. Downloading...")

        with tqdm(unit="B", unit_scale=True, miniters=1, desc="Downloading") as t:
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

            urllib.request.urlretrieve(
                "https://zenodo.org/record/1188976/files/Audio_Speech_Actors_01-24.zip",
                ravdess_archive,
                reporthook,
            )

        with zipfile.ZipFile(ravdess_archive, "r") as archive:
            members = archive.namelist()
            for member in tqdm(members, desc="Extracting RAVDESS dataset"):
                archive.extract(member, ravdess_dir)

        return ravdess_dir

    except Exception as e:
        print("Error downloading or extracting RAVDESS dataset:", e)

        cleanup([ravdess_dir])
        return None

    finally:
        if os.path.isfile(ravdess_archive):
            try:
                os.remove(ravdess_archive)
            except Exception as e:
                print(f"Failed to remove archive {ravdess_archive}: {e}")

def collate_fn_train(batch, feature_extractor):
    audio = [example["audio"]["array"] for example in batch]
    labels = torch.tensor([example["label"] for example in batch])

    features = feature_extractor(
        audio,
        sampling_rate=16000,
        padding=True,
        return_tensor="pt",
    )

    features["labels"] = labels
    return {
      "input_values": features["input_values"],
      "attention_mask": features["attention_mask"],
      "labels": labels,
    }

def compute_metrics(eval_pred):
    logits, labels = eval_pred 
    preds = np.argmax(logits, axis=-1)
    acc = accuracy.compute(predictions=preds, references=labels)["accuracy"]

    return {"accuracy": acc}

def main():
    ravdess_dir = os.path.join(SCRIPT_DIR, "ravdess")
    ravdess_archive = os.path.join(SCRIPT_DIR, "ravdess.zip")

    ravdess_path = load_ravdess_dataset(ravdess_archive, ravdess_dir)
    if ravdess_path is None:
        return

    emotion_dict = {
        "01": "neutral",
        "02": "calm",
        "03": "happy",
        "04": "sad",
        "05": "angry",
        "06": "fearful",
        "07": "disgust",
        "08": "surprised",
    }

    emotions = list(emotion_dict.values())
    emotion2id = {e: i for i, e in enumerate(emotions)}

    files = glob.glob(os.path.join(ravdess_path, "**", "*.wav"), recursive=True)

    valid_paths = []
    labels = []

    for path in files:
        fname = os.path.basename(path)
        parts = fname.split("-")

        if len(parts) != 7:
            continue

        emotion = emotion_dict.get(parts[2])
        if emotion is None:
            continue

        valid_paths.append(path)
        labels.append(emotion2id[emotion])

    dataset = Dataset.from_dict({"audio": valid_paths, "label": labels})

    dataset = dataset.cast_column("audio", Audio(sampling_rate=16000))

    split = dataset.train_test_split(test_size=0.2, seed=67)
    dataset = {"train": split["train"], "test": split["test"]}

    val_split = dataset["train"].train_test_split(test_size=0.1, seed=67)
    dataset["train"] = val_split["train"]
    dataset["validation"] = val_split["test"]

    training_args = TrainingArguments(
        per_device_train_batch_size = 4,
        num_train_epochs = 5,
        learning_rate = 1e-5,
        logging_steps = 20,
        logging_strategy = "epoch",
        save_strategy = "epoch",
        fp16=True,
        remove_unused_columns=False,
        eval_strategy = "epoch",
        metric_for_best_model = "eval_accuracy",
        load_best_model_at_end=True,
        report_to="none"
    )
    

if __name__ == "__main__":
    main()
