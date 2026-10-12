(() => {
  const root = document.getElementById("pg");
  if (!root) return;
  const $ = (id) => document.getElementById(id);

  const el = {
    opSelect: $("pg-op"),
    opHelp: $("pg-op-help"),
    params: $("pg-params"),
    common: $("pg-common"),
    opMode: $("pg-op-mode"),
    pipeMode: $("pg-pipe-mode"),
    pipeline: $("pg-pipeline"),
    preset: $("pg-preset"),
    examples: $("pg-examples"),
    seed: $("pg-seed"),
    random: $("pg-random"),
    auto: $("pg-auto"),
    run: $("pg-run"),
    file: $("pg-file"),
    sample: $("pg-sample"),
    size: $("pg-size"),
    compare: $("pg-compare"),
    download: $("pg-download"),
    stage: $("pg-stage"),
    canvas: $("pg-canvas"),
    original: $("pg-original"),
    status: $("pg-status"),
    command: $("pg-command"),
    copy: $("pg-copy"),
  };
  const ctx = el.canvas.getContext("2d");
  const tabs = root.querySelectorAll("[data-mode]");

  const examplePipelines = [
    "sort angle=270 interval=edges edge=0.15 | shift | scanlines",
    "resize scale=0.2 | sort angle=90 lower=0.1 | resize original filter=nearest",
    "sort path=flow interval=none",
    "@cinematic | glitch slices=12",
    "sort key=hue lower=0.1 | gradientmap colors=vapor | mix 0.5",
    "neon | blend mode=screen",
  ];

  let info = null;
  let mode = "op";
  let controls = [];
  let imageReady = false;
  let runId = 0;
  let busy = false;
  let pending = null;
  let debounce = null;
  let hasResult = false;
  let initialState = readHash();

  const worker = new Worker("worker.js");
  worker.onmessage = (e) => {
    const m = e.data;
    if (m.type === "fatal") {
      setStatus("Could not start the playground: " + m.error, "error");
    } else if (m.type === "ready") {
      info = m.info;
      setup();
    } else if (m.type === "image") {
      if (m.error) {
        setStatus(m.error, "error");
        return;
      }
      imageReady = true;
      const scaled = m.width !== m.originalWidth ? ` (scaled down from ${m.originalWidth}×${m.originalHeight})` : "";
      setStatus(`Image loaded: ${m.width}×${m.height}${scaled}`);
      schedule(true);
    } else if (m.type === "result") {
      busy = false;
      if (m.id === runId) {
        if (m.error) {
          setStatus(m.error, "error");
        } else {
          el.canvas.width = m.width;
          el.canvas.height = m.height;
          ctx.putImageData(new ImageData(new Uint8ClampedArray(m.pixels), m.width, m.height), 0, 0);
          hasResult = true;
          el.stage.classList.add("pg-has-result");
          setStatus(`${m.width}×${m.height} · ${m.ms} ms`);
        }
      }
      if (pending) {
        const p = pending;
        pending = null;
        send(p.id, p.req);
      }
    }
  };

  function setStatus(text, kind) {
    el.status.textContent = text;
    el.status.classList.toggle("pg-error", kind === "error");
  }

  function readHash() {
    const params = new URLSearchParams(location.hash.slice(1));
    if (![...params.keys()].length) return null;
    return {
      mode: params.get("mode") || "op",
      op: params.get("op"),
      args: params.getAll("a"),
      pipeline: params.get("p"),
      seed: params.get("seed"),
    };
  }

  function writeHash(req) {
    const params = new URLSearchParams();
    params.set("mode", mode);
    if (mode === "op") {
      params.set("op", req.op);
      req.args.forEach((a) => params.append("a", a));
    } else {
      params.set("p", req.pipeline);
    }
    params.set("seed", String(req.seed));
    history.replaceState(null, "", "#" + params.toString());
  }

  function setup() {
    for (const op of info.ops) {
      const o = document.createElement("option");
      o.value = op.name;
      o.textContent = op.name;
      el.opSelect.appendChild(o);
    }
    for (const name of Object.keys(info.presets).sort()) {
      const o = document.createElement("option");
      o.value = name;
      o.textContent = "@" + name;
      o.title = info.presets[name];
      el.preset.appendChild(o);
    }
    for (const p of examplePipelines) {
      const b = document.createElement("button");
      b.type = "button";
      b.className = "pg-chip";
      b.textContent = p;
      b.addEventListener("click", () => {
        el.pipeline.value = p;
        schedule(true);
      });
      el.examples.appendChild(b);
    }

    const start = initialState || { mode: "op", op: "sort", args: ["angle=270", "interval=edges", "edge=0.15"], pipeline: examplePipelines[0], seed: "7" };
    el.seed.value = start.seed || "7";
    el.pipeline.value = start.pipeline || examplePipelines[0];
    el.opSelect.value = info.ops.some((o) => o.name === start.op) ? start.op : "sort";
    buildParams(start.args || []);
    setMode(start.mode === "pipeline" ? "pipeline" : "op", false);
    initialState = null;

    tabs.forEach((t) => t.addEventListener("click", () => setMode(t.dataset.mode, true)));
    el.opSelect.addEventListener("change", () => {
      buildParams([]);
      schedule(true);
    });
    el.preset.addEventListener("change", () => {
      if (!el.preset.value) return;
      el.pipeline.value = "@" + el.preset.value;
      el.preset.value = "";
      schedule(true);
    });
    el.pipeline.addEventListener("input", () => schedule());
    el.seed.addEventListener("input", () => schedule());
    el.random.addEventListener("click", () => {
      el.seed.value = String(Math.floor(Math.random() * 1000000));
      schedule(true);
    });
    el.auto.addEventListener("change", () => el.auto.checked && schedule(true));
    el.run.addEventListener("click", () => schedule(true, true));
    el.file.addEventListener("change", () => el.file.files[0] && loadFile(el.file.files[0]));
    el.sample.addEventListener("click", loadSample);
    el.size.addEventListener("change", () => (currentSource ? currentSource() : loadSample()));
    el.copy.addEventListener("click", () => navigator.clipboard && navigator.clipboard.writeText(el.command.textContent));
    el.download.addEventListener("click", download);

    const show = (on) => el.stage.classList.toggle("pg-show-original", on && hasResult);
    el.compare.addEventListener("pointerdown", () => show(true));
    for (const ev of ["pointerup", "pointerleave", "pointercancel"]) el.compare.addEventListener(ev, () => show(false));

    el.stage.addEventListener("dragover", (e) => {
      e.preventDefault();
      el.stage.classList.add("pg-dragging");
    });
    el.stage.addEventListener("dragleave", () => el.stage.classList.remove("pg-dragging"));
    el.stage.addEventListener("drop", (e) => {
      e.preventDefault();
      el.stage.classList.remove("pg-dragging");
      const f = e.dataTransfer.files[0];
      if (f) loadFile(f);
    });

    loadSample();
  }

  let currentSource = null;

  function sendImage(buffer) {
    imageReady = false;
    setStatus("Loading image…");
    worker.postMessage({ type: "image", bytes: buffer, maxSide: Number(el.size.value) }, [buffer]);
  }

  function loadSample() {
    currentSource = loadSample;
    el.original.src = "city.jpg";
    fetch("city.jpg")
      .then((r) => r.arrayBuffer())
      .then(sendImage)
      .catch((err) => setStatus("Could not load the sample photo: " + err, "error"));
  }

  function loadFile(file) {
    currentSource = () => loadFile(file);
    el.original.src = URL.createObjectURL(file);
    file.arrayBuffer().then(sendImage);
  }

  function setMode(m, run) {
    mode = m;
    tabs.forEach((t) => t.classList.toggle("pg-active", t.dataset.mode === m));
    el.opMode.hidden = m !== "op";
    el.pipeMode.hidden = m !== "pipeline";
    if (run) schedule(true);
  }

  function control(p, value) {
    const row = document.createElement("label");
    row.className = "pg-field";
    const name = document.createElement("span");
    name.textContent = p.name;
    row.appendChild(name);
    let input;
    if (p.kind === "bool") {
      row.className = "pg-check";
      input = document.createElement("input");
      input.type = "checkbox";
      input.checked = (value ?? p.default) === "true";
      row.insertBefore(input, name);
    } else if (p.choices && p.choices.length) {
      input = document.createElement("select");
      const choices = p.choices.includes(p.default) || !p.default ? p.choices : [p.default, ...p.choices];
      if (!p.default) choices.unshift("");
      for (const c of choices) {
        const o = document.createElement("option");
        o.value = c;
        o.textContent = c || "(none)";
        input.appendChild(o);
      }
      input.value = value ?? p.default;
      row.appendChild(input);
    } else {
      input = document.createElement("input");
      input.type = p.kind === "int" || p.kind === "float" ? "number" : "text";
      if (p.kind === "float") input.step = "any";
      input.value = value ?? p.default;
      input.placeholder = p.default;
      row.appendChild(input);
    }
    row.title = p.help;
    const help = document.createElement("small");
    help.textContent = p.help;
    row.appendChild(help);
    input.addEventListener(p.kind === "bool" || input.tagName === "SELECT" ? "change" : "input", () => schedule());
    return { p, input, row };
  }

  function buildParams(args) {
    const op = info.ops.find((o) => o.name === el.opSelect.value);
    const given = {};
    for (const a of args) {
      const i = a.indexOf("=");
      if (i > 0) given[a.slice(0, i)] = a.slice(i + 1);
    }
    el.opHelp.textContent = op.help;
    el.params.textContent = "";
    el.common.textContent = "";
    controls = [];
    for (const p of op.params) {
      if (p.name === "seed") continue;
      const c = control(p, given[p.name]);
      controls.push(c);
      el.params.appendChild(c.row);
    }
    for (const p of info.common) {
      if (op.params.some((q) => q.name === p.name)) continue;
      const c = control(p, given[p.name]);
      controls.push(c);
      el.common.appendChild(c.row);
    }
    if (!op.params.length) {
      const none = document.createElement("p");
      none.className = "pg-help";
      none.textContent = "This operation has no parameters.";
      el.params.appendChild(none);
    }
  }

  function changedArgs() {
    const args = [];
    for (const { p, input } of controls) {
      const v = p.kind === "bool" ? String(input.checked) : input.value.trim();
      const same = p.kind === "int" || p.kind === "float" ? Number(v) === Number(p.default) && v !== "" : v === p.default;
      if (!same && !(v === "" && p.default === "")) args.push(`${p.name}=${v}`);
    }
    return args;
  }

  function seed() {
    const n = Number.parseInt(el.seed.value, 10);
    return Number.isFinite(n) ? n : 0;
  }

  function request() {
    if (mode === "pipeline") return { pipeline: el.pipeline.value.trim(), seed: seed() };
    return { op: el.opSelect.value, args: changedArgs(), seed: seed() };
  }

  function quote(v) {
    return /^[\w.,:#@/+=-]+$/.test(v) ? v : `"${v.replace(/"/g, '\\"')}"`;
  }

  function command(req) {
    if (req.pipeline !== undefined) return `sortmancy run -i in.jpg -o out.png --seed ${req.seed} ${quote(req.pipeline)}`;
    const flags = req.args.map((a) => {
      const i = a.indexOf("=");
      const k = a.slice(0, i);
      const v = a.slice(i + 1);
      return v === "true" ? `--${k}` : `--${k} ${quote(v)}`;
    });
    return ["sortmancy", req.op, "-i in.jpg -o out.png", ...flags, `--seed ${req.seed}`].join(" ");
  }

  function schedule(now, force) {
    if (!info) return;
    const req = request();
    el.command.textContent = command(req);
    writeHash(req);
    if (!imageReady) return;
    if (mode === "pipeline" && !req.pipeline) return;
    if (!el.auto.checked && !force) return;
    clearTimeout(debounce);
    debounce = setTimeout(() => {
      runId++;
      if (busy) pending = { id: runId, req };
      else send(runId, req);
    }, now ? 0 : 250);
  }

  function send(id, req) {
    busy = true;
    setStatus("Rendering…");
    worker.postMessage({ type: "run", id, request: req });
  }

  function download() {
    if (!hasResult) return;
    el.canvas.toBlob((blob) => {
      const a = document.createElement("a");
      a.href = URL.createObjectURL(blob);
      a.download = "sortmancy.png";
      a.click();
      setTimeout(() => URL.revokeObjectURL(a.href), 1000);
    }, "image/png");
  }
})();
