function generateStartOverlay(overlayElement, ws, gameId) {
    overlayElement.innerHTML = startGameTemplate().trim();
    addStartButtonListener(ws, gameId);
    addCopyEventListener(gameId);
    overlayElement.classList.remove('hidden');
}

function generateWaitOverlay(overlayElement, gameId) {
    overlayElement.innerHTML = joinGameTemplate();
    addCopyEventListener(gameId);
    overlayElement.classList.remove('hidden');
}

function generateGameCompleteOverlay(overlayElement, playerNameScrHtmlElemMap) {
    playerNameScrHtmlElemMap = new Map(Object.entries(playerNameScrHtmlElemMap));
    overlayElement.innerHTML = completeGameTemplate();
    overlayElement.classList.remove('hidden');
    for(const elem of playerNameScrHtmlElemMap.values()) {
        elem.classList.add("mx-auto");
        document.getElementById("st").appendChild(elem);
    }
}

function hideOverlayElement(overlayElement) {
    overlayElement.classList.add('hidden');
}

function startGameTemplate() {
   return  `<div class="bg-white p-8 rounded-lg shadow-lg">
                <div class="mx-4">
                    <button id="start_game" class="mt-4 mx-auto bg-red-500 text-white px-4 py-2 rounded">Click To Start</button>
                </div>
                <div class="mx-4">
                    <button id="copy_game_id" class="mt-4 mx-auto bg-red-500 text-white px-4 py-2 rounded">Click To Copy GameId</button>
                </div>
            </div>`;
}

function joinGameTemplate() {
   return   `<div class="bg-white p-8 rounded-lg shadow-lg">
                <div class="mx-4">
                    <h2 class="text-2xl font-bold mb-4">Waiting for game to start ...</h2>
                </div>
                <div class="mx-4">
                    <button id="copy_game_id" class="mt-4 mx-auto bg-red-500 text-white px-4 py-2 rounded">Click To Copy GameId</button>
                </div>
            </div>`;
}

function completeGameTemplate() {
    return `<div class="bg-white p-8 rounded-lg shadow-lg">
                <h2 class="text-3xl text-center font-bold text-red-500">
                    GAME OVER
                </h2>
                <div id="st" class="flex flex-col lg:flex-row mx-4">
                </div>
            </div>`;
}


function addStartButtonListener(ws, gameId) {
    document.getElementById("start_game").addEventListener("click", () => {
        ws.send(JSON.stringify({
            Action: "start",
            GameId: gameId
        }))});
}

function addCopyEventListener(gameId) {
    document.getElementById("copy_game_id").addEventListener("click", () => {
        copyGameId(document.getElementById("copy_game_id"), gameId)
    })
}

function copyGameId(copyGameElem, gameId){
  
  const tempElem = document.createElement("textarea");
  tempElem.value = gameId;
  document.body.appendChild(tempElem);
  tempElem.select();
  tempElem.setSelectionRange(0, 99999); // For mobile devices
  document.execCommand('copy');
  document.body.removeChild(tempElem);
  
  let removedClasses = [];
  copyGameElem.classList.forEach( classname => {
    if(classname.startsWith("hover:") || (classname.startsWith("bg-"))){
      copyGameElem.classList.remove(classname);
      removedClasses.push(classname);
    }
  });

  copyGameElem.classList.add("bg-green-100");
  setTimeout(() => {
    copyGameElem.classList.remove("bg-green-100");
    removedClasses.forEach((className) => {
      copyGameElem.classList.add(className);
    })
    copyGameElem.textContent = "Click To Copy GameId";
  }, 1000)
}


export{generateStartOverlay,
    generateWaitOverlay,
    generateGameCompleteOverlay,
    hideOverlayElement};