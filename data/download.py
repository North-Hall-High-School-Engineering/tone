import os
import shutil
import tarfile
import urllib.request
from datetime import datetime

from tqdm import tqdm

DATASETS = {
    "MELD": "https://huggingface.co/datasets/declare-lab/MELD/resolve/main/MELD.Raw.tar.gz"
}

artifacts = []


class ProgressBar(tqdm):
    def update_to(self, block_num, block_size, total_size):
        if total_size > 0:
            self.total = total_size
        self.update(block_num * block_size - self.n)


def download(url: str, output_path: str):
    with ProgressBar(
        unit="B", unit_scale=True, miniters=1, desc=os.path.basename(output_path)
    ) as t:
        urllib.request.urlretrieve(url, filename=output_path, reporthook=t.update_to)


def cleanup(paths):
    for path in paths:
        if os.path.exists(path):
            if os.path.isfile(path):
                os.remove(path)
                print(f"Deleted file: {path}")
            elif os.path.isdir(path):
                shutil.rmtree(path)
                print(f"Deleted directory: {path}")


def extract_tar_gz(file_path, extract_to):
    try:
        with tarfile.open(file_path, "r:gz") as tar:
            tar.extractall(path=extract_to)
            print(f"Extracted {file_path} to {extract_to}")
            artifacts.append(file_path)

            for root, _, files in os.walk(extract_to):
                for f in files:
                    if f.endswith((".tar.gz", ".tgz")):
                        nested_path = os.path.join(root, f)
                        nested_extract_dir = os.path.join(root, f.split(".")[0])
                        os.makedirs(nested_extract_dir, exist_ok=True)
                        extract_tar_gz(nested_path, nested_extract_dir)
                        os.remove(nested_path)

    except Exception as e:
        print(f"Failed to extract {file_path}: {e}")
        if os.path.exists(file_path):
            os.remove(file_path)


def main():
    data_dir = os.path.join(os.path.dirname(os.path.abspath(__file__)), "raw")
    os.makedirs(data_dir, exist_ok=True)

    for name, url in DATASETS.items():
        dataset_dir = os.path.join(data_dir, name)
        output_file = os.path.join(data_dir, f"{name}.tar.gz")
        start = datetime.now()
        print(f"Processing dataset {name} started at {start}")

        try:
            download(url, output_file)
            end = datetime.now()
            print(f"Finished downloading {name} at {end}, duration: {end - start}")

            os.makedirs(dataset_dir, exist_ok=True)
            extract_tar_gz(output_file, dataset_dir)

        except Exception as e:
            print(f"Failed to download or extract {name}: {e}")
            if os.path.exists(output_file):
                print(f"Deleting incomplete file {output_file}")
                os.remove(output_file)

    cleanup(artifacts)


if __name__ == "__main__":
    main()
