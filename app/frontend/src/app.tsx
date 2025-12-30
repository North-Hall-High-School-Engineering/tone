import { h, Fragment } from "preact";
import { useEffect, useState } from "preact/hooks";
import { Orb } from "./components/orb";
import { EventsOn } from "../wailsjs/runtime";

export function App(props: any) {
  const [audioDebug, setAudioDebug] = useState({
    numSamples: 0,
    maxValue: 0,
    minValue: 0,
  });

  useEffect(() => {
    EventsOn("audioDebug", (event: any) => {
      setAudioDebug(event);
    });
  }, []);

  return (
    <>
      <div class="w-screen h-screen flex flex-col items-center">
        <div className="w-full h-1/2">
          <Orb displacementStrength={0.5} noiseIntensity={0.125} />
        </div>
        <div className="w-full h-full items-start p-4">
          <h3>Audio Debug Info:</h3>
          <p>Num Samples: {audioDebug.numSamples}</p>
          <p>Max Value: {audioDebug.maxValue.toFixed(3)}</p>
          <p>Min Value: {audioDebug.minValue.toFixed(3)}</p>
        </div>
      </div>
    </>
  );
}
