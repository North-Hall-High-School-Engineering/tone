import { h, Fragment } from "preact";
import { useEffect, useState } from "preact/hooks";
import { Orb } from "./components/orb";
import { EventsOn } from "../wailsjs/runtime";

export function App(props: any) {
  const [preds, setPreds] = useState<number[]>([0, 0, 0, 0, 0, 0, 0, 0, 0]);

  useEffect(() => {
    EventsOn("inference", setPreds);
  }, []);

  return (
    <>
      <div class="w-screen h-screen flex flex-col items-center">
        <div className="w-full h-1/2">
          <Orb displacementStrength={0.5} noiseIntensity={0.125} />
        </div>
      </div>
    </>
  );
}
