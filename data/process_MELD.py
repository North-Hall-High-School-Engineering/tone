import os
import shutil
import subprocess
from concurrent.futures import ProcessPoolExecutor, as_completed


def does_raw_exist(meld_path: str):
    try:
        must_exist = [
            meld_path,
            *[os.path.join(meld_path, subdir) for subdir in ["dev", "test", "train"]],
        ]
        for path in must_exist:
            if not os.path.exists(path):
                print(f"Required path does not exist: {path}")
                return False
        return True
    except PermissionError as e:
        print(f"Permission denied: {e}")
        return False
    except Exception as e:
        print(f"Unexpected error: {e}")
        return False


def collect(parent_path, allowed_exts=None):
    collected_files = []
    if allowed_exts is not None:
        allowed_exts = set(allowed_exts)
    for root, dirs, files in os.walk(parent_path):
        for file in files:
            if file.startswith("._"):
                continue
            full_path = os.path.join(root, file)
            if allowed_exts:
                _, ext = os.path.splitext(file)
                if ext.lower() not in allowed_exts:
                    continue
            collected_files.append(full_path)
    return collected_files


def convert_mp4_to_wav_task(args):
    mp4_path, wav_path = args
    os.makedirs(os.path.dirname(wav_path), exist_ok=True)

    cmd = [
        "ffmpeg",
        "-i",
        mp4_path,
        "-vn",
        "-acodec",
        "pcm_s16le",
        "-ar",
        "16000",
        "-y",
        wav_path,
    ]

    try:
        subprocess.run(
            cmd, check=True, stdout=subprocess.DEVNULL, stderr=subprocess.PIPE
        )
        return f"Converted {mp4_path} -> {wav_path}"
    except subprocess.CalledProcessError as e:
        return f"Failed to convert {mp4_path}: {e.stderr.decode()}"


def process_split_parallel(split_name, files, processed_dir, max_workers=None):
    audio_dir = os.path.join(processed_dir, split_name, "audio")
    tasks = []

    for file_path in files:
        file_name = os.path.splitext(os.path.basename(file_path))[0]
        wav_path = os.path.join(audio_dir, f"{file_name}.wav")
        tasks.append((file_path, wav_path))

    with ProcessPoolExecutor(max_workers=max_workers) as executor:
        futures = [executor.submit(convert_mp4_to_wav_task, task) for task in tasks]
        for future in as_completed(futures):
            print(future.result())


def copy_csv_files(meld_path, processed_dir):
    csv_map = {
        "train": os.path.join(meld_path, "train", "train_sent_emo.csv"),
        "dev": os.path.join(meld_path, "dev_sent_emo.csv"),
        "test": os.path.join(meld_path, "test_sent_emo.csv"),
    }

    for split, src_csv in csv_map.items():
        dst_dir = os.path.join(processed_dir, split)
        os.makedirs(dst_dir, exist_ok=True)

        if os.path.exists(src_csv):
            shutil.copy(src_csv, dst_dir)
            print(f"Copied CSV for {split}: {src_csv} -> {dst_dir}")
        else:
            print(f"CSV for {split} not found: {src_csv}")


def main():
    raw_dir = os.path.join(os.path.dirname(os.path.abspath(__file__)), "raw")
    meld_path = os.path.join(raw_dir, "MELD", "MELD.Raw")

    if not does_raw_exist(meld_path):
        print("does_raw_exist() returned False, terminating.")
        return 1

    processed_dir = os.path.join(
        os.path.dirname(os.path.abspath(__file__)), "processed", "MELD"
    )

    splits = {
        "train": collect(
            os.path.join(meld_path, "train", "train_splits"), allowed_exts={".mp4"}
        ),
        "test": collect(
            os.path.join(meld_path, "test", "output_repeated_splits_test"),
            allowed_exts={".mp4"},
        ),
        "dev": collect(
            os.path.join(meld_path, "dev", "dev_splits_complete"), allowed_exts={".mp4"}
        ),
    }

    for split_name, files in splits.items():
        process_split_parallel(split_name, files, processed_dir)

    copy_csv_files(meld_path, processed_dir)

    print("All files converted successfully.")


if __name__ == "__main__":
    main()
