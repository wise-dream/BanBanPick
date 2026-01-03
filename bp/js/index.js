const finalOverlay = document.getElementById("finalOverlay");
  const finalMapNameEl = document.getElementById("finalMapName");
  const closeFinalBtn = document.getElementById("closeFinalBtn");

  const sideOverlay = document.getElementById("sideOverlay");
  const sideResult = document.getElementById("sideResult");
  const sideSubtitleEl = document.getElementById("sideSubtitle");
  const closeSideBtn = document.getElementById("closeSideBtn");
  const sideFinalCard = document.querySelector(".side-final-card");

  const sideFullOverlay = document.getElementById("sideFullOverlay");
  const sideFullText = document.getElementById("sideFullText");
  const sideFullIcon = document.getElementById("sideFullIcon");

  const allMaps = [
    "Ascent","Bind","Haven","Split","Lotus","Icebox",
    "Sunset","Pearl","Fracture","Breeze","Corrode","Abyss"
  ];

  const mapsGrid = document.getElementById("mapsGrid");
  const teamANameInput = document.getElementById("teamAName");
  const teamBNameInput = document.getElementById("teamBName");
  const teamABox = document.getElementById("teamABox");
  const teamBBox = document.getElementById("teamBBox");
  const stepLabel = document.getElementById("stepLabel");
  const resetBtn = document.getElementById("resetBtn");
  const swapBtn = document.getElementById("swapBtn");
  const startBtn = document.getElementById("startBtn");
  const sideBtn = document.getElementById("sideBtn");
  const logBox = document.getElementById("logBox");
  const summaryContent = document.getElementById("summaryContent");

  let state = {
    currentTeam: "A",
    bans: [],
    pickedMap: null,
    finished: false,
    started: false
  };

  const BAN_TIME_LIMIT = 20;
  let timerId = null;
  let timeLeft = BAN_TIME_LIMIT;

  function updateActiveTeam() {
    teamABox.classList.toggle("active", state.currentTeam === "A");
    teamBBox.classList.toggle("active", state.currentTeam === "B");
  }

  function updateStepLabelWithTimer() {
    const teamName =
      state.currentTeam === "A" ? teamANameInput.value : teamBNameInput.value;
    stepLabel.textContent = `Ход бана: ${teamName} | ${timeLeft}s`;
  }

  function startBanTimer() {
    clearInterval(timerId);
    timeLeft = BAN_TIME_LIMIT;
    updateStepLabelWithTimer();

    timerId = setInterval(() => {
      timeLeft -= 1;
      if (timeLeft <= 0) {
        clearInterval(timerId);
        state.currentTeam = state.currentTeam === "A" ? "B" : "A";
        updateActiveTeam();
        timeLeft = BAN_TIME_LIMIT;
      }
      updateStepLabelWithTimer();
    }, 1000);
  }

  function stopBanTimer() {
    clearInterval(timerId);
  }

  function resetState() {
    state = {
      currentTeam: "A",
      bans: [],
      pickedMap: null,
      finished: false,
      started: false
    };
    logBox.innerHTML = "";
    summaryContent.textContent =
      "Пока карта не выбрана. Баньте, пока не останется одна.";
    stepLabel.textContent = "Нажмите «Начать», чтобы начать вето";
    stepLabel.classList.remove("done");
    stepLabel.classList.add("step");
    updateActiveTeam();
    renderMaps();
    stopBanTimer();
    finalOverlay.style.display = "none";
    sideOverlay.style.display = "none";
    sideFullOverlay.style.display = "none";
  }

  function startVeto() {
    if (state.started || state.finished) return;
    state.started = true;
    log("Новый Bo1 с банами до последней карты начат.");
    startBanTimer();
  }

  function swapCurrentTeam() {
    if (!state.started || state.finished) return;
    state.currentTeam = state.currentTeam === "A" ? "B" : "A";
    updateActiveTeam();
    const teamName =
      state.currentTeam === "A" ? teamANameInput.value : teamBNameInput.value;
    log(`Ход передан команде <strong>${teamName}</strong>.`);
    startBanTimer();
  }

  function log(message) {
    const div = document.createElement("div");
    div.className = "log-entry";
    div.innerHTML = message;
    logBox.appendChild(div);
    logBox.scrollTop = logBox.scrollHeight;
  }

  function onBan(mapName) {
    if (!state.started || state.finished) return;
    if (state.bans.includes(mapName)) return;

    stopBanTimer();

    const teamName =
      state.currentTeam === "A" ? teamANameInput.value : teamBNameInput.value;
    state.bans.push(mapName);
    log(`<strong>${teamName}</strong> банит карту <strong>${mapName}</strong>.`);

    const remaining = allMaps.filter(m => !state.bans.includes(m));

    if (remaining.length === 1) {
      const lastMap = remaining[0];
      state.pickedMap = lastMap;
      state.finished = true;
      log(
        `<strong>Автопик:</strong> последняя оставшаяся карта <strong>${lastMap}</strong> выбирается для игры.`
      );
      summaryContent.textContent = `Играется карта: ${lastMap}. Сторону выбираете при заходе в лобби.`;
      stepLabel.textContent = "Veto завершён";
      stepLabel.classList.remove("step");
      stepLabel.classList.add("done");
      showFinalMap(lastMap);
    } else {
      state.currentTeam = state.currentTeam === "A" ? "B" : "A";
      updateActiveTeam();
      startBanTimer();
    }

    renderMaps();
  }

  function renderMaps() {
    mapsGrid.innerHTML = "";
    allMaps.forEach(map => {
      const isBanned = state.bans.includes(map);
      const isPicked = state.pickedMap === map;

      const card = document.createElement("div");
      card.className = "map-card";

      const mapClass = "map-" + map.toLowerCase();
      card.classList.add(mapClass.replace(/\s+/g, "-"));

      if (isBanned || isPicked || (state.finished && !isPicked)) {
        card.classList.add("map-disabled");
      }

      const name = document.createElement("div");
      name.className = "map-name";
      name.textContent = map;

      const status = document.createElement("div");
      status.className = "map-status";
      if (isPicked) status.textContent = "Выбрана для игры";
      else if (isBanned) status.textContent = "Забанена";
      else status.textContent = "Доступна для бана";

      const actions = document.createElement("div");
      actions.className = "map-actions";

      if (!isBanned && !isPicked && !state.finished) {
        const banBtn = document.createElement("button");
        banBtn.className = "btn btn-accent";
        banBtn.textContent = "Ban";
        banBtn.onclick = () => onBan(map);
        actions.appendChild(banBtn);
      }

      card.appendChild(name);
      card.appendChild(status);
      card.appendChild(actions);

      if (isBanned || isPicked) {
        const tag = document.createElement("div");
        tag.className = "map-tag " + (isPicked ? "picked" : "banned");
        tag.textContent = isPicked ? "PICKED" : "BANNED";
        card.appendChild(tag);
      }

      mapsGrid.appendChild(card);
    });
  }

  function showFinalMap(mapName) {
    finalMapNameEl.textContent = mapName;
    const urlName = mapName.toLowerCase().replace(/\s+/g, "-");
    const imgUrl = `img/${urlName}.png`;
    const card = finalOverlay.querySelector(".final-card");
    card.style.setProperty("--final-bg", `url("${imgUrl}")`);
    finalOverlay.style.display = "flex";
  }

  function showSideRandom() {
    if (!state.finished || !state.pickedMap) {
      alert("Сначала завершите вето и выберите карту.");
      return;
    }

    const sides = ["ATTACK", "DEFENCE"];

    sideResult.textContent = "ATTACK / DEFENCE";
    sideSubtitleEl.textContent = "Выбор стороны...";

    sideFullText.textContent = "ATTACK / DEFENCE";
    sideFullIcon.src = "img/attack.png";
    sideFullOverlay.style.display = "flex";

    let ticks = 0;
    const maxTicks = 30;

    function spinStep() {
      const randomSide = sides[Math.floor(Math.random() * sides.length)];
      sideFullText.textContent = randomSide;
      sideFullIcon.src =
        randomSide === "ATTACK" ? "img/attack.png" : "img/defence.png";
      ticks++;

      if (ticks < maxTicks) {
        const progress = ticks / maxTicks;
        const minDelay = 40;
        const maxDelay = 250;
        const delay =
          minDelay + (maxDelay - minDelay) * progress * progress;

        setTimeout(spinStep, delay);
      } else {
        const finalSide = sides[Math.floor(Math.random() * sides.length)];
        const teamA = teamANameInput.value || "Team A";
        const teamB = teamBNameInput.value || "Team B";
        const otherSide = finalSide === "ATTACK" ? "DEFENCE" : "ATTACK";

        sideFullOverlay.style.display = "none";

        sideResult.textContent = finalSide;
        sideSubtitleEl.textContent =
          `${teamA} — ${finalSide}, ${teamB} — ${otherSide}`;

        const iconPath =
          finalSide === "ATTACK" ? "img/attack.png" : "img/defence.png";

        sideFinalCard.style.setProperty("--final-bg", `url("${iconPath}")`);

        sideOverlay.style.display = "flex";
      }
    }

    spinStep();
  }

  closeFinalBtn.addEventListener("click", () => {
    finalOverlay.style.display = "none";
  });

  closeSideBtn.addEventListener("click", () => {
    sideOverlay.style.display = "none";
  });

  resetBtn.addEventListener("click", resetState);
  swapBtn.addEventListener("click", swapCurrentTeam);
  startBtn.addEventListener("click", startVeto);
  sideBtn.addEventListener("click", showSideRandom);

  resetState();