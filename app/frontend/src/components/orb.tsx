import { useEffect, useRef } from "preact/hooks";
import * as THREE from "three";
import { h } from "preact";
import { createNoise3D } from "simplex-noise";

export type OrbProps = {
  displacementStrength: number;
  noiseIntensity: number;
  color: number;
};

export function Orb({ displacementStrength, noiseIntensity, color }: OrbProps) {
  const mountRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    if (!mountRef.current) return;

    const scene = new THREE.Scene();
    // scene.background = new THREE.Color(0x101010);

    const camera = new THREE.PerspectiveCamera(
      45,
      mountRef.current.clientWidth / mountRef.current.clientHeight,
      0.1,
      1000,
    );
    camera.position.z = 3;

    const renderer = new THREE.WebGLRenderer({ antialias: true, alpha: true });
    renderer.setSize(
      mountRef.current.clientWidth,
      mountRef.current.clientHeight,
    );
    mountRef.current.appendChild(renderer.domElement);

    const lightTop = new THREE.DirectionalLight(0xffffff, 0.7);
    lightTop.position.set(0, 5, 5);
    scene.add(lightTop);

    const lightBottom = new THREE.DirectionalLight(0xffffff, 0.25);
    lightBottom.position.set(0, -5, 5);
    scene.add(lightBottom);

    const ambientLight = new THREE.AmbientLight(0x798296);
    scene.add(ambientLight);

    const geometry = new THREE.SphereGeometry(1, 128, 128);
    const material = new THREE.MeshPhongMaterial({
      color: color,
      shininess: 100,
    });
    const sphere = new THREE.Mesh(geometry, material);
    scene.add(sphere);

    const originalPositions = geometry.attributes.position.array.slice();

    const simplex = createNoise3D();

    const animate = () => {
      const time = performance.now() * 0.0005;

      const vertices = geometry.attributes.position;
      for (let i = 0; i < vertices.count; i++) {
        const x = originalPositions[i * 3 + 0];
        const y = originalPositions[i * 3 + 1];
        const z = originalPositions[i * 3 + 2];

        const noise =
          simplex(
            x * displacementStrength,
            y * displacementStrength,
            z * displacementStrength + time,
          ) * noiseIntensity;
        const scale = 1 + noise;

        vertices.setXYZ(i, x * scale, y * scale, z * scale);
      }

      vertices.needsUpdate = true;
      geometry.computeVertexNormals();

      sphere.rotation.y += 0.003;
      renderer.render(scene, camera);
      requestAnimationFrame(animate);
    };

    animate();

    return () => {
      if (!mountRef.current) return;
      mountRef.current.removeChild(renderer.domElement);
      renderer.dispose();
    };
  }, [displacementStrength, noiseIntensity]);

  return <div ref={mountRef} style={{ width: "100%", height: "100%" }} />;
}
