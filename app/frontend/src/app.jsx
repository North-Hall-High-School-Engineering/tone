import { h } from "preact";
import { Orb } from "./components/orb";
import { useEffect, useState, useRef } from "preact/hooks";
import { EventsOn } from "../wailsjs/runtime";

export function App(props) {
  const [loudness, setLoudness] = useState(0);
  const smoothLoudnessRef = useRef(0);
  const meanLoudnessRef = useRef(0.001);
  const SMOOTHING = 0.05;
  const MEAN_SMOOTHING = 0.01;

  useEffect(() => {
    const loudnessCb = EventsOn("loudness", (rms) => {
      smoothLoudnessRef.current =
        smoothLoudnessRef.current * (1 - SMOOTHING) + rms * SMOOTHING;

      meanLoudnessRef.current =
        meanLoudnessRef.current * (1 - MEAN_SMOOTHING) +
        smoothLoudnessRef.current * MEAN_SMOOTHING;

      const normalized = smoothLoudnessRef.current / meanLoudnessRef.current;

      setLoudness(normalized);
    });

    return () => {
      loudnessCb();
    };
  }, []);

  const noiseScale = 0.08 * loudness;

  return (
    <div className="w-screen h-screen">
      <div className="w-full h-2/3">
        <Orb
          displacementStrength={1}
          noiseIntensity={noiseScale}
          color={0xe4ecfa}
        />
      </div>
    </div>
  );
}
