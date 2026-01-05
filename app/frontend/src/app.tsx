import { h, Fragment } from "preact";
import { useEffect, useState } from "preact/hooks";
import { Orb } from "./components/orb";
import { EventsOn } from "../wailsjs/runtime";

const EMOTION_LABELS = [
  "frustrated",
  "angry",
  "sad",
  "disgust",
  "excited",
  "fear",
  "neutral",
  "surprise",
  "happy",
];

export function App(props: any) {
  const [preds, setPreds] = useState<number[]>(
    new Array(EMOTION_LABELS.length).fill(0),
  );
  const [vad, setVad] = useState<number>(0);

  useEffect(() => {
    const offVad = EventsOn("vad", (p: number) => {
      setVad(p);
    });

    const offInf = EventsOn("inference", (values: number[]) => {
      setPreds(values);
    });

    return () => {
      offVad();
      offInf();
    };
  }, []);

  return (
    <div class="w-screen h-screen flex flex-col items-center p-4">
      {/* Orb */}
      <div className="w-full h-1/2 mb-4">
        <Orb
          displacementStrength={0.5}
          noiseIntensity={0.125}
          color={0xe4ecfa}
        />
      </div>

      {/* VAD */}
      <div class="w-full mb-4">
        <div class="text-sm font-medium">VAD: {vad.toFixed(2)}</div>
        <div
          class="h-2 bg-green-400 rounded transition-all"
          style={{ width: `${Math.min(vad * 100, 100)}%` }}
        />
      </div>

      {/* SER Predictions */}
      <div class="w-full space-y-1">
        {preds.map((p, i) => (
          <div key={i} class="flex items-center space-x-2">
            <div class="w-24 text-xs">{EMOTION_LABELS[i]}</div>
            <div class="flex-1 h-2 bg-gray-300 rounded overflow-hidden">
              <div
                class="h-2 bg-blue-500 rounded transition-all"
                style={{ width: `${Math.min(p * 100, 100)}%` }}
              />
            </div>
            <div class="w-8 text-xs text-right">{(p * 100).toFixed(0)}%</div>
          </div>
        ))}
      </div>
    </div>
  );
}
