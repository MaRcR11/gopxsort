importScripts("wasm_exec.js");

const go = new Go();

self.sortmancyReady = () => {
  postMessage({ type: "ready", info: JSON.parse(self.sortmancy.info()) });
};

fetch("sortmancy.wasm")
  .then((r) => {
    if (!r.ok) throw new Error("could not load sortmancy.wasm (HTTP " + r.status + ")");
    return r.arrayBuffer();
  })
  .then((buf) => WebAssembly.instantiate(buf, go.importObject))
  .then((res) => go.run(res.instance))
  .catch((err) => postMessage({ type: "fatal", error: String(err) }));

onmessage = (e) => {
  const m = e.data;
  try {
    if (m.type === "image") {
      const r = self.sortmancy.setImage(new Uint8Array(m.bytes), m.maxSide);
      postMessage({
        type: "image",
        error: r.error || null,
        width: r.width,
        height: r.height,
        originalWidth: r.originalWidth,
        originalHeight: r.originalHeight,
      });
    } else if (m.type === "run") {
      const r = self.sortmancy.run(JSON.stringify(m.request));
      if (r.error) {
        postMessage({ type: "result", id: m.id, error: r.error });
        return;
      }
      const buf = r.pixels.buffer;
      postMessage({ type: "result", id: m.id, width: r.width, height: r.height, ms: r.ms, pixels: buf }, [buf]);
    }
  } catch (err) {
    postMessage({ type: m.type === "image" ? "image" : "result", id: m.id, error: String(err) });
  }
};
