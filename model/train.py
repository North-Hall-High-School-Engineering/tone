import glob
import os
import urllib.request
import zipfile

from datasets import Dataset

# !wget -q https://zenodo.org/record/1188976/files/Audio_Speech_Actors_01-24.zip
# !unzip -q Audio_Speech_Actors_01-24.zip -d ravdess


def main():
    ravdess_fname = "ravdess.zip"
    if not os.path.exists(ravdess_fname):
        urllib.request.urlretrieve(
            "https://zenodo.org/record/1188976/files/Audio_Speech_Actors_01-24.zip",
            ravdess_fname,
        )

        with zipfile.ZipFile(ravdess_fname, "r") as archive:
            archive.extractall("ravdess")

    emotion_dict = {
        "01": "neutral",
        "02": "calm",
        "03": "happy",
        "04": "sad",
        "05": "angry",
        "06": "fearful",
        "07": "disgust",
        "08": "surprised",
    }  # code -> label

    emotions = list(emotion_dict.values())
    emotion2id = {e: i for i, e in enumerate(emotions)}

    files = glob.glob(
        os.path.join(ravdess_fname, "/**/*.wav")
    )  # collect all .wav files

    valid_paths = []
    labels = []

    for path in files:
        fname = os.path.basename(path)
        parts = fname.split("-")

        if len(parts) != 7:
            print("invalid filename, skipping", fname)
            continue

        code = parts[2]
        emotion = emotion_dict[code]

        if not emotion:
            print("invalid emotion, skipping", fname)
            continue

        valid_paths.append(path)
        labels.append(emotion2id[emotion])

    dataset = Dataset.from_dict({"audio": valid_paths, "label": labels})


if __name__ == "__main__":
    main()
