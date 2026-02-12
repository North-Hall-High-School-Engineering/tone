import { h } from "preact";
import { useEffect, useRef } from "preact/hooks";
import * as THREE from "three";
import { createNoise3D } from "simplex-noise";

export function Orb({ displacementStrength, noiseIntensity, color }) {
  const mountRef = useRef(null);

  // Three.js refs
  const sceneRef = useRef();
  const cameraRef = useRef();
  const rendererRef = useRef();
  const sphereRef = useRef();
  const geometryRef = useRef();
  const materialRef = useRef();
  const originalPositionsRef = useRef();
  const simplexRef = useRef();

  // Prop refs to hold latest values
  const displacementStrengthRef = useRef(displacementStrength);
  const noiseIntensityRef = useRef(noiseIntensity);
  const colorRef = useRef(color);

  // Update refs whenever props change
  useEffect(() => {
    displacementStrengthRef.current = displacementStrength;
  }, [displacementStrength]);

  useEffect(() => {
    noiseIntensityRef.current = noiseIntensity;
  }, [noiseIntensity]);

  useEffect(() => {
    colorRef.current = color;
    if (materialRef.current) {
      materialRef.current.color.set(color);
    }
  }, [color]);

  useEffect(() => {
    if (!mountRef.current) return;

    // Scene
    const scene = new THREE.Scene();
    sceneRef.current = scene;

    // Camera
    const camera = new THREE.PerspectiveCamera(
      45,
      mountRef.current.clientWidth / mountRef.current.clientHeight,
      0.1,
      1000,
    );
    camera.position.z = 3;
    cameraRef.current = camera;

    // Renderer
    const renderer = new THREE.WebGLRenderer({ antialias: true, alpha: true });
    renderer.setSize(
      mountRef.current.clientWidth,
      mountRef.current.clientHeight,
    );
    mountRef.current.appendChild(renderer.domElement);
    rendererRef.current = renderer;

    // Lights
    const lightTop = new THREE.DirectionalLight(0xffffff, 0.7);
    lightTop.position.set(0, 5, 5);
    scene.add(lightTop);

    const lightBottom = new THREE.DirectionalLight(0xffffff, 0.25);
    lightBottom.position.set(0, -5, 5);
    scene.add(lightBottom);

    const ambientLight = new THREE.AmbientLight(0x798296);
    scene.add(ambientLight);

    // Sphere
    const geometry = new THREE.SphereGeometry(1, 128, 128);
    geometryRef.current = geometry;

    const material = new THREE.MeshPhongMaterial({
      color: colorRef.current,
      shininess: 100,
    });
    materialRef.current = material;

    const sphere = new THREE.Mesh(geometry, material);
    scene.add(sphere);
    sphereRef.current = sphere;

    originalPositionsRef.current = geometry.attributes.position.array.slice();
    simplexRef.current = createNoise3D();

    let frameId;

    const animate = () => {
      if (!geometryRef.current || !sphereRef.current || !simplexRef.current)
        return;

      const time = performance.now() * 0.0005;
      const vertices = geometryRef.current.attributes.position;

      const ds = displacementStrengthRef.current;
      const ni = noiseIntensityRef.current;

      for (let i = 0; i < vertices.count; i++) {
        const x = originalPositionsRef.current[i * 3 + 0];
        const y = originalPositionsRef.current[i * 3 + 1];
        const z = originalPositionsRef.current[i * 3 + 2];

        const noise = simplexRef.current(x * ds, y * ds, z * ds + time) * ni;
        const scale = 1 + noise;

        vertices.setXYZ(i, x * scale, y * scale, z * scale);
      }

      vertices.needsUpdate = true;
      geometryRef.current.computeVertexNormals();

      sphereRef.current.rotation.y += 0.003;
      renderer.render(scene, camera);

      frameId = requestAnimationFrame(animate);
    };

    animate();

    return () => {
      cancelAnimationFrame(frameId);
      if (!mountRef.current) return;
      mountRef.current.removeChild(renderer.domElement);
      renderer.dispose();
    };
  }, []);

  return <div ref={mountRef} style={{ width: "100%", height: "100%" }} />;
}
