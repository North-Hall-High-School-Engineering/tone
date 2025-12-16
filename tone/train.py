from functools import partial

import numpy as np
import torch
from datasets import Audio, load_dataset
from transformers import (
    EarlyStoppingCallback,
    HubertModel,
    Trainer,
    TrainingArguments,
    Wav2Vec2FeatureExtractor,
)

from models.combined import CombinedModel


def collate_fn_train(batch, feature_extractor):
    audio_arrays = [item["audio"]["array"] for item in batch]

    features = feature_extractor(
        audio_arrays,
        sampling_rate=16000,
        padding=True,
        return_tensors="pt",
        return_attention_mask=True,
    )

    labels = torch.tensor(
        [
            [
                item["frustrated"],
                item["angry"],
                item["sad"],
                item["disgust"],
                item["excited"],
                item["fear"],
                item["neutral"],
                item["surprise"],
                item["happy"],
            ]
            for item in batch
        ],
        dtype=torch.float32,
    )

    return {
        "input_values": features["input_values"],
        "attention_mask": features["attention_mask"],
        "labels": labels,
    }


def compute_metrics(eval_pred):
    predictions, labels = eval_pred

    predictions = np.asarray(predictions, dtype=np.float32)
    labels = np.asarray(labels, dtype=np.float32)

    mse = np.mean((predictions - labels) ** 2)
    return {"mse": mse}


if __name__ == "__main__":
    if torch.cuda.is_available():
        print("Using GPU")
    else:
        print(
            "Using CPU\nPlease refer to SETUP.md to set up CUDA/GPU for faster training"
        )

    dataset = load_dataset("AbstractTTS/IEMOCAP")
    dataset = dataset.cast_column("audio", Audio(sampling_rate=16000))

    split = dataset["train"].train_test_split(test_size=0.2, seed=67)
    dataset = {"train": split["train"], "test": split["test"]}

    val_split = dataset["train"].train_test_split(test_size=0.1, seed=67)
    dataset["train"] = val_split["train"]
    dataset["validation"] = val_split["test"]

    feature_extractor = Wav2Vec2FeatureExtractor.from_pretrained(
        "facebook/hubert-base-ls960"
    )

    hubert = HubertModel.from_pretrained("facebook/hubert-base-ls960")
    for param in hubert.parameters():
        param.requires_grad = False

    model = CombinedModel(hubert)

    training_args = TrainingArguments(
        output_dir="./checkpoints",
        per_device_train_batch_size=8,
        num_train_epochs=20,
        logging_steps=10,
        logging_strategy="steps",
        save_strategy="epoch",
        fp16=torch.cuda.is_available(),
        report_to="none",
        remove_unused_columns=False,
        dataloader_num_workers=4,
        dataloader_pin_memory=True,
        load_best_model_at_end=True,
        eval_strategy="epoch",
        metric_for_best_model="eval_loss",
    )

    trainer = Trainer(
        model=model,
        args=training_args,
        train_dataset=dataset["train"],
        data_collator=partial(collate_fn_train, feature_extractor=feature_extractor),
        callbacks=[
            EarlyStoppingCallback(
                early_stopping_patience=2, early_stopping_threshold=1e-4
            )
        ],
        eval_dataset=dataset["validation"],
        compute_metrics=compute_metrics,
    )

    trainer.train()

    trainer.save_model(".")
    test_metrics = trainer.evaluate(dataset["test"])
    print(test_metrics)

    inference_model = trainer.model.model
    inference_model.cpu()
    inference_model.eval()

    dummy_input_values = torch.randn(1, 16000)  # 1 second @ 16kHz
    dummy_attention_mask = torch.ones(1, 16000, dtype=torch.long)

    torch.onnx.export(
        model=inference_model,
        args=(dummy_input_values, dummy_attention_mask),
        f="tone_ser.onnx",
        input_names=["input_values", "attention_mask"],
        output_names=["preds"],
        dynamic_axes={
            "input_values": {0: "batch", 1: "time"},
            "attention_mask": {0: "batch", 1: "time"},
            "preds": {0: "batch"},
        },
        opset_version=17,
        do_constant_folding=True,
    )
