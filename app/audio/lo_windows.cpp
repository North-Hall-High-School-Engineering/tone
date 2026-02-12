#include "lo_windows.h"
#include <AudioSessionTypes.h>
#include <audioclient.h>
#include <handleapi.h>
#include <initguid.h>
#include <combaseapi.h>
#include <stdio.h>
#include <stdlib.h>
#include <avrt.h>
#include <synchapi.h>
#include <winbase.h>
#include <string.h>

struct Loopback {
    IAudioClient* audioClient = nullptr;
    IAudioCaptureClient* captureClient = nullptr;
    WAVEFORMATEX* pwfx = nullptr;

    HANDLE hEvent = nullptr;
    HANDLE hTask = nullptr;
};

extern "C" Loopback* loopback_start() {
    HRESULT hr;

    hr = CoInitialize(NULL);
    if (FAILED(hr)) return nullptr;

    Loopback* lb = new Loopback();

    IMMDeviceEnumerator* enumerator = nullptr;
    IMMDevice* device = nullptr;
    DWORD tidx = 0;

    hr = CoCreateInstance(__uuidof(MMDeviceEnumerator), nullptr, CLSCTX_ALL,
                          __uuidof(IMMDeviceEnumerator), (void**)&enumerator);
    if (FAILED(hr)) goto fail;

    hr = enumerator->GetDefaultAudioEndpoint(eRender, eConsole, &device);
    if (FAILED(hr)) goto fail;

    hr = device->Activate(__uuidof(IAudioClient), CLSCTX_ALL, nullptr, (void**)&lb->audioClient);
    if (FAILED(hr)) goto fail;

    hr = lb->audioClient->GetMixFormat(&lb->pwfx);
    if (FAILED(hr)) goto fail;

    lb->hEvent = CreateEvent(NULL, FALSE, FALSE, NULL);
    if (!lb->hEvent) goto fail;

    hr = lb->audioClient->Initialize(AUDCLNT_SHAREMODE_SHARED,
                                     AUDCLNT_STREAMFLAGS_LOOPBACK | AUDCLNT_STREAMFLAGS_EVENTCALLBACK,
                                     0, 0, lb->pwfx, nullptr);
    if (FAILED(hr)) goto fail;

    hr = lb->audioClient->SetEventHandle(lb->hEvent);
    if (FAILED(hr)) goto fail;

    hr = lb->audioClient->GetService(__uuidof(IAudioCaptureClient), (void**)&lb->captureClient);
    if (FAILED(hr)) goto fail;

    lb->hTask = AvSetMmThreadCharacteristicsW(L"Pro Audio", &tidx);
    if (!lb->hTask) goto fail;

    hr = lb->audioClient->Start();
    if (FAILED(hr)) goto fail;

    if (device) device->Release();
    if (enumerator) enumerator->Release();

    return lb;

fail:
    if (lb->captureClient) lb->captureClient->Release();
    if (lb->audioClient) lb->audioClient->Release();
    if (lb->pwfx) CoTaskMemFree(lb->pwfx);
    if (lb->hEvent) CloseHandle(lb->hEvent);
    delete lb;
    if (device) device->Release();
    if (enumerator) enumerator->Release();
    CoUninitialize();
    return nullptr;
}

extern "C" int loopback_read(Loopback* lb, BYTE** data, UINT32* numFrames) {
    if (!lb || !lb->captureClient || !lb->pwfx) return -1;

    DWORD wait = WaitForSingleObject(lb->hEvent, INFINITE);
    if (wait != WAIT_OBJECT_0) return -1;

    UINT32 packetLength = 0;
    HRESULT hr = lb->captureClient->GetNextPacketSize(&packetLength);
    if (FAILED(hr)) return -1;
    if (packetLength == 0) return 0;

    BYTE* pData = nullptr;
    UINT32 frames = 0;
    DWORD flags;

    hr = lb->captureClient->GetBuffer(&pData, &frames, &flags, nullptr, nullptr);
    if (FAILED(hr)) return -1;

    UINT32 bytes = frames * lb->pwfx->nBlockAlign;
    *data = (BYTE*)malloc(bytes);
    if (!*data) {
        lb->captureClient->ReleaseBuffer(frames);
        return -1;
    }

    if (flags & AUDCLNT_BUFFERFLAGS_SILENT)
        memset(*data, 0, bytes);
    else
        memcpy(*data, pData, bytes);

    *numFrames = frames;

    lb->captureClient->ReleaseBuffer(frames);

    return 1;
}

extern "C" void loopback_stop(Loopback* lb) {
    if (!lb) return;

    if (lb->audioClient) lb->audioClient->Stop();
    if (lb->hTask) AvRevertMmThreadCharacteristics(lb->hTask);
    if (lb->captureClient) lb->captureClient->Release();
    if (lb->audioClient) lb->audioClient->Release();
    if (lb->pwfx) CoTaskMemFree(lb->pwfx);
    if (lb->hEvent) CloseHandle(lb->hEvent);

    delete lb;
    CoUninitialize();
}

extern "C" WAVEFORMATEX* loopback_get_waveformat(Loopback* lb) {
    if (!lb) return nullptr;
    return lb->pwfx;
}
