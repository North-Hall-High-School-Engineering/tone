import glob
import os
import shutil
import urllib.request
import zipfile

import evaluate
import numpy as np
import torch
from audiomentations import AddGaussianNoise, Compose, TimeStretch
from dataset_util import *
from datasets import Audio, Dataset
from tqdm import tqdm
from transformers import (
    AutoFeatureExtractor,
    AutoModel,
    DataCollatorWithPadding,
    EarlyStoppingCallback,
    Trainer,
    TrainingArguments,
    WavLMForSequenceClassification,
)

SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
CACHE_DIR = os.path.join(SCRIPT_DIR, ".cache")
OUT_DIR = os.path.join(SCRIPT_DIR, "out")

MODEL_NAME = "microsoft/wavlm-large"
BATCH_SIZE = 32
EPOCHS = 6
LR = 2e-5
MAX_SAMPLES = 16000 * 5
N_PROC = os.cpu_count() or 1

accuracy = evaluate.load("accuracy")

os.makedirs(CACHE_DIR, exist_ok=True)

feature_extractor = AutoFeatureExtractor.from_pretrained(MODEL_NAME)

augment = Compose(
    [
        AddGaussianNoise(min_amplitude=0.001, max_amplitude=0.010, p=0.25),
        TimeStretch(min_rate=0.9, max_rate=1.1, p=0.25),
    ]
)


def compute_metrics(eval_pred):
    logits, labels = eval_pred
    preds = np.argmax(logits, axis=-1)
    acc = accuracy.compute(predictions=preds, references=labels)["accuracy"]

    return {"accuracy": acc}


def augment_waveform(wav, sr):
    return augment(samples=wav, sample_rate=sr)


def preprocess(batch, train=False):
    audio = batch["audio"]
    wav = audio["array"]
    sr = audio["sampling_rate"]

    if len(wav) > MAX_SAMPLES:
        wav = wav[:MAX_SAMPLES]

    if train:
        wav = augment_waveform(wav, sr)

    if len(wav) > MAX_SAMPLES:
        wav = wav[:MAX_SAMPLES]

    inputs = feature_extractor(
        wav,
        sampling_rate=sr,
        padding=False,
        return_attention_mask=True,
    )

    batch["input_values"] = inputs["input_values"][0]
    batch["attention_mask"] = inputs["attention_mask"][0]
    return batch


def main():
    base_dir = "data"

    ravdess_dir = download_ravdess(base_dir)
    tess_dir = download_tess(base_dir)
    cremad_dir = download_cremad(base_dir)
    emodb_dir = download_emodb(base_dir)

    dataset, emotion2id = load_all_datasets(
        ravdess_dir=ravdess_dir,
        tess_dir=tess_dir,
        cremad_dir=cremad_dir,
        emodb_dir=emodb_dir,
    )

    split = dataset.train_test_split(test_size=0.2, seed=67)
    val_split = split["train"].train_test_split(test_size=0.1, seed=67)

    dataset = {
        "train": val_split["train"],
        "validation": val_split["test"],
        "test": split["test"],
    }

    dataset["train"] = dataset["train"].map(
        lambda x: preprocess(x, train=True),
        num_proc=N_PROC,
        remove_columns=["audio"],
    )

    dataset["validation"] = dataset["validation"].map(
        lambda x: preprocess(x, train=False),
        num_proc=N_PROC,
        remove_columns=["audio"],
    )

    dataset["test"] = dataset["test"].map(
        lambda x: preprocess(x, train=False),
        num_proc=N_PROC,
        remove_columns=["audio"],
    )

    model = WavLMForSequenceClassification.from_pretrained(
        MODEL_NAME,
        num_labels=len(emotion2id),
        problem_type="single_label_classification",
    )

    training_args = TrainingArguments(
        output_dir="./wavlm_checkpoints",
        eval_strategy="epoch",
        save_strategy="epoch",
        learning_rate=LR,
        per_device_train_batch_size=BATCH_SIZE,
        per_device_eval_batch_size=BATCH_SIZE,
        num_train_epochs=EPOCHS,
        warmup_ratio=0.1,
        max_grad_norm=1.0,
        logging_steps=25,
        save_total_limit=2,
        load_best_model_at_end=True,
        metric_for_best_model="accuracy",
        report_to="none",
        remove_unused_columns=False,
        dataloader_pin_memory=True,
    )

    data_collator = DataCollatorWithPadding(
        feature_extractor, padding=True, return_tensors="pt"
    )

    trainer = Trainer(
        model=model,
        args=training_args,
        train_dataset=dataset["train"],
        eval_dataset=dataset["validation"],
        compute_metrics=compute_metrics,
        data_collator=data_collator,
        callbacks=[EarlyStoppingCallback(early_stopping_patience=3)],
    )

    trainer.train()
    trainer.save_model(OUT_DIR)
    feature_extractor.save_pretrained(OUT_DIR)


if __name__ == "__main__":
    main()
