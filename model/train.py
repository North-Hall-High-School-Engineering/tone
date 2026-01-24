import glob
import os
import urllib.request
import zipfile

from datasets import Dataset
from tqdm import tqdm


# !wget -q https://zenodo.org/record/1188976/files/Audio_Speech_Actors_01-24.zip
# !unzip -q Audio_Speech_Actors_01-24.zip -d ravdess

def load_ravdess_dataset():
    ravdess_fname = "ravdess.zip"
    try:
        if not os.path.exists(ravdess_fname):
            print("Ravdess dataset not found. Downloading...")

            with tqdm(unit="B", unit_scale=True, miniters=1, desc="Progress: ") as t:
                def reporthook(block_size, total_size):
                    if t.total > 0:
                        t.total = total_size
                    t.update(block_size)
                urllib.request.urlretrieve(
                    "https://zenodo.org/record/1188976/files/Audio_Speech_Actors_01-24.zip",
                    ravdess_fname,
                    reporthook,
                )

        with zipfile.ZipFile(ravdess_fname, "r") as archive:
            members = archive.namelist()

            for member in tqdm(members, desc="Extracting RAVDESS dataset"):
                archive.extract(member, "ravdess")    
        print("RAVDESS dataset downloaded and extracted.")
        return "ravdess"
    
    except Exception as e:
        print("Error downloading or extracting RAVDESS dataset:", e)
        return None


def main():
    ravdess_fname = load_ravdess_dataset()
    if ravdess_fname is None:
        print("Failed to load RAVDESS dataset.")
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
    }  # code -> label

    emotions = list(emotion_dict.values())
    emotion2id = {e: i for i, e in enumerate(emotions)}

    files = glob.glob(
    os.path.join(ravdess_fname, "**", "*.wav"),
    recursive=True
)

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
    print(dataset)



if __name__ == "__main__":
    main()
