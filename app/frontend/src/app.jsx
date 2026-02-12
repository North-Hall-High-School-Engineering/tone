import { h } from "preact";
import { Orb } from "./components/orb";
import { useEffect, useState, useRef } from "preact/hooks";
import { EventsOn } from "../wailsjs/runtime";

const EMOTION_STYLE = {
  angry: {
    color: 0xef4444,
    noise: 1.0,
    pulse: 2.2,
    glow: 1.4,
  },
  disgust: {
    color: 0x65a30d,
    noise: 0.8,
    pulse: 0.6,
    glow: 0.8,
  },
  fearful: {
    color: 0x7c3aed,
    noise: 0.6,
    pulse: 1.8,
    glow: 1.1,
  },
  happy: {
    color: 0xfbbf24,
    noise: 0.8,
    pulse: 1.6,
    glow: 1.6,
  },
  neutral: {
    color: 0x94a3b8,
    noise: 0.4,
    pulse: 0.5,
    glow: 0.6,
  },
  sad: {
    color: 0x3b82f6,
    noise: 0.3,
    pulse: 0.4,
    glow: 0.5,
  },
};

export function App(props) {
  const [loudness, setLoudness] = useState(0);
  const [predictions, setPredictions] = useState([]);
  const [topLabel, setTopLabel] = useState(null);
  const smoothLoudnessRef = useRef(0);
  const meanLoudnessRef = useRef(0.001);
  const SMOOTHING = 0.05;
  const MEAN_SMOOTHING = 0.01;

  useEffect(() => {
    const loudnessCb = EventsOn("loudness", (rms) => {
      if (rms<.00001){
        return;
      }
      smoothLoudnessRef.current =
        smoothLoudnessRef.current * (1 - SMOOTHING) + rms * SMOOTHING;

      meanLoudnessRef.current =
        meanLoudnessRef.current * (1 - MEAN_SMOOTHING) +
        smoothLoudnessRef.current * MEAN_SMOOTHING;

      const normalized = smoothLoudnessRef.current / meanLoudnessRef.current;

      setLoudness(normalized);
    });

    const inferenceCb = EventsOn("inference", (preds) => {
      if (!Array.isArray(preds)) return;

      const sorted = [...preds].sort((a, b) => b.score - a.score);

      setPredictions(sorted);

      if (sorted.length > 0) {
        setTopLabel(sorted[0]);
      }
    });

    return () => {
      loudnessCb();
      inferenceCb();
    };
  }, []);

  const confidence = topLabel?.score ?? 0;
  const label = topLabel?.label ?? "neutral";

  const style = EMOTION_STYLE[label] ?? EMOTION_STYLE.neutral;

  const noiseScale = 0.09 * loudness * style.noise;

  const life = Math.min(1, confidence * 1.5);

  const pulseSpeed = (0.5 + confidence * 2) * style.pulse;

  const glowStrength = life * style.glow;

  return (
    <div className="w-screen h-screen flex flex-col items-center">
      <div id="titlebar"></div>
      <div className="w-full h-1/2">
        <Orb
          displacementStrength={1 + life * 0.3}
          noiseIntensity={noiseScale}
          pulseSpeed={pulseSpeed}
          glowStrength={glowStrength}
          color={style.color}
        />
      </div>
    </div>
  );
}
