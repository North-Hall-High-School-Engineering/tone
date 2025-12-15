3.11 <= Python Version <= 3.12.12 required due to a breaking change in the requests module bundled with 🤗 transformers.

To install Python 3.12, the recommended version for this project, visit https://www.python.org/downloads/release/python-3120/

Make sure you get the installer and not the python manager as that will automatically install the newest version.
Windows (x64): https://www.python.org/ftp/python/3.12.0/python-3.12.0-amd64.exe

Verify installation
```bash
  py --list
```

You should have an output similar to this showing version 3.12 as an avalible option if installation was successful
```
-V:3.14 *        Python 3.14 (64-bit)
-V:3.12          Python 3.12 (64-bit)
-V:3.10          Python 3.10 (64-bit)
```

Create a venv with Python 3.12 as the primary version
```bash
  py -3.12 -m venv .venv
```

Activate the venv
Windows: `.\.venv\Scripts\activate`
Mac/Linux: `source .venv/bin/activate`

Install deps inside tone dir
```
  pip install -r requirements.txt
```

If running on an NVIDIA GPU, make sure to install CUDA from https://developer.nvidia.com/cuda-downloads

If you want to run the training on your gpu, run these commands in the venv
```bash
  pip uninstall torch torchaudio torchvision
```

```bash
  pip install torch torchaudio torchvision --index-url https://download.pytorch.org/whl/cu118
```

Common Issues

"Torch not compiled with CUDA enabled"
Install CUDA from https://developer.nvidia.com/cuda-downloads or switch to cpu device within torch. If you have already installed cuda, just restart your terminal and the problem will likely be fixed. If the issue persists, restart your computer and/or re-install cuda.

RuntimeError: Could not load libtorchcodec. Likely causes:
          1. FFmpeg is not properly installed in your environment. We support
             versions 4, 5, 6, 7, and 8. On Windows, ensure you've installed
             the "full-shared" version which ships DLLs.
          2. The PyTorch version (2.7.1+cu118) is not compatible with
             this version of TorchCodec. Refer to the version compatibility
             table:
             https://github.com/pytorch/torchcodec?tab=readme-ov-file#installing-torchcodec.
          3. Another runtime dependency; see exceptions below.
        The following exceptions were raised as we tried to load libtorchcodec:
        
The cause of this issue is 🤗 datasets module being >= 4.0.
```bash
    pip install datasets<4.0.0
```
https://discuss.huggingface.co/t/issue-with-torchcodec-when-fine-tuning-whisper-asr-model/169315/2
