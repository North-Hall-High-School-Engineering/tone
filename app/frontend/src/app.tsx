import { h, Fragment } from "preact";
import { useEffect, useState } from "preact/hooks";
import { Orb } from "./components/orb";

export function App(props: any) {
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
