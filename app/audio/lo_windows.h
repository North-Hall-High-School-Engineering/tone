#ifndef LOOPBACK_WINDOWS_H
#define LOOPBACK_WINDOWS_H

#include <Windows.h>
#include <audioclient.h>
#include <mmdeviceapi.h>

#ifdef __cplusplus
extern "C" {
#endif

typedef struct Loopback Loopback;

Loopback* loopback_start();
int loopback_read(Loopback *lb, BYTE **data, UINT32 *numFrames);
void loopback_stop(Loopback *lb);
WAVEFORMATEX* loopback_get_waveformat(Loopback* lb);

#ifdef __cplusplus
}
#endif

#endif // LOOPBACK_WINDOWS_H
