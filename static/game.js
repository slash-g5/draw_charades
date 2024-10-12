import rough from "https://cdn.jsdelivr.net/npm/roughjs@4.3.1/bundled/rough.esm.js";
import { draw } from "./roughjsHelper/RoughCanvasDraw.js";
import { wsHost, httpHost } from "./config/uiConfig.js";
import { loadThemeRecursive} from "./theme/theme.js";
import { loadPlayerByConnectionId, reloadGameStateFromServer } from "./gameRefresher/gameResourceLoader.js";
import { hideOverlayElement, generateWaitOverlay, generateStartOverlay, generateGameCompleteOverlay} from "./overlay/overlayGenerator.js";

// Default theme
let theme = "purple"

const ws = new WebSocket(wsHost);
const urlParams = new URLSearchParams(window.location.search);
const name = urlParams.get("name");
const mode = urlParams.get("mode");
let gameId = urlParams.get("gameId");

const imgElementClass = "h-16 lg:h-24 w:4d";
const nameElemClass = "text-xl font-bold p-2";
const scoreElemClass = "screlem px-2 pb-2 font-bold";

const indvScoreBaseClass = `flex flex-row border-2 border-` + theme + `-400` +  ` bg-` + theme + `-100 mt-4 lg:mt-8 rounded-lg lg:p-2 truncate w-52 lg:w-80`;

if(mode == null || !(mode === "create" || mode === "join") ||
   name == null || name === "" || 
   (mode === "join" && (gameId == null || gameId === ""))){
  window.location.href = httpHost + "/welcome?err=unable to join"
}

const avatar = urlParams.get("avatar");
let conId = null

//gameId display
let gameIdDisplay = document.getElementById("game_id_display");

//Full game state
let gameState = null

//Chat window
const msgList = document.getElementById("msg_list");
const chatButton = document.getElementById("chat_button");
const chatInput = document.getElementById("chat_input");

//Score List
const scoreSheet = document.getElementById("score_sheet");
const playerNameScrHtmlElemMap = {}
const playerConIdNameMap = {}
const rankingNames = []

//Title element
const titleElement = document.getElementById("title");

//Overlay element
const overlay = document.getElementById("overlay");

//Canvas related
let drawer = false;
let isDrawing = false;
let drawItem = "pencil";
let drawColor = "black";
let canvaColor = "#FFFFFF";
let drawing = [];
let undoWindow = [];
let currentShape = [];

//Canvas draw tools
const pencilTool = document.getElementById("pencil_tool");
const lineTool = document.getElementById("line_tool");
const rectTool = document.getElementById("rect_tool");
const brushTool = document.getElementById("brush_tool");
const clearTool = document.getElementById("clear_tool");
const eraserTool = document.getElementById("eraser_tool");
const undoTool = document.getElementById("undo_tool");
const redoTool = document.getElementById("redo_tool");

const gameCanvas = document.getElementById("game_canvas");
const roughCanvas = rough.canvas(gameCanvas);
const canvasContext = gameCanvas.getContext("2d");

updateTheme()

gameCanvas.setAttribute(
  "width",
  Math.min(gameCanvas.parentNode.offsetWidth, 600),
);

gameCanvas.setAttribute(
  "height",
  Math.min(gameCanvas.parentNode.offsetWidth, 600),
);

let canvasRect = gameCanvas.getBoundingClientRect();
const canvasWidth = canvasRect.width;
const canvasHeight = canvasRect.height;
const canvasTopLeft = {
  top: canvasRect.top,
  left: canvasRect.left,
  scaleX: 0,
  scaleY: 0,
};

//gameCanvas Event Listeners
{
  gameCanvas.addEventListener("mousemove", (e) => {
    updateCanvasCordinates();
    if (!drawer) {
      return;
    }
    if (!isDrawing) {
      return;
    }
    updateDrawing(e);
    updateCanvas();
  });

  gameCanvas.addEventListener("touchmove", (eTouch) => {
    updateCanvasCordinates();

    if (!drawer) {
      return;
    }
    eTouch.preventDefault();
    let e = {
      "x": eTouch.touches[0].clientX,
      "y": eTouch.touches[0].clientY,
    };
    updateDrawing(e);
    updateCanvas();
  });

  gameCanvas.addEventListener("mousedown", (_) => {
    if (drawer) {
      isDrawing = true;
    }
  });

  gameCanvas.addEventListener("mouseup", (_) => {
    isDrawing = false;
    currentShape = [];
  });

  gameCanvas.addEventListener("mouseleave", (_) => {
    isDrawing = false;
    currentShape = [];
  });

  gameCanvas.addEventListener("touchend", (_) => {
    isDrawing = false;
    currentShape = [];
  });
}

//drawTools Event Listeners
{
  pencilTool.addEventListener("click", () => {
    if (!drawer) {
      return;
    }
    drawItem = "pencil";
  });

  lineTool.addEventListener("click", () => {
    if (!drawer) {
      return;
    }
    drawItem = "line";
  });

  rectTool.addEventListener("click", () => {
    if (!drawer) {
      return;
    }
    drawItem = "rect";
  });

  brushTool.addEventListener("click", () => {
    if (!drawer) {
      return;
    }
    drawItem = "brush";
  });

  clearTool.addEventListener("click", () => {
    if (!drawer) {
      return;
    }
    updateDrawingAndClear();
  });

  eraserTool.addEventListener("click", () => {
    if (!drawer) {
      return;
    }
    drawItem = "eraser";
  });

  undoTool.addEventListener("click", () => {
    if (!drawer) {
      return;
    }
    performUndo();
  });

  redoTool.addEventListener("click", () => {
    if (!drawer) {
      return;
    }
    performRedo();
  });
}

//gameIdDisplay Event Listeners
gameIdDisplay.addEventListener("click", () => {
  copyGameIdToClipboard();
})

//Theme Button Event Listeners
{
  redTheme.addEventListener("click", () => {
    theme = 'red';
    updateTheme();
  })

  yellowTheme.addEventListener("click", () => {
    theme = 'yellow';
    updateTheme();
  })

  blueTheme.addEventListener("click", () => {
    theme = 'blue';
    updateTheme();
  })

  greenTheme.addEventListener("click", () => {
    theme = 'green';
    updateTheme();
  })

  cyanTheme.addEventListener("click", () => {
    theme = 'cyan';
    updateTheme();
  })

  tealTheme.addEventListener("click", () => {
    theme = 'teal';
    updateTheme();
  })

  pinkTheme.addEventListener("click", () => {
    theme = 'pink';
    updateTheme();
  })

  purpleTheme.addEventListener("click", () => {
    theme = 'purple';
    updateTheme();
  })
}

//Chat Event Listeners
chatButton.addEventListener("click", () => {
  if(!chatInput.value || chatInput.value.length > 25){
    return;
  }
  ws.send(JSON.stringify({
    Action: "chat",
    ClientId: conId,
    GameId: gameId,
    ChatText: chatInput.value
  }))
})

//websocket message Actions
{
  ws.onmessage = (message) => {
    console.log(message.data);
    const response = JSON.parse(message.data);
    if (response.Action === "connect") {
      conId = response.ConnectionId;
      createGameIfNeeded();
      joinGameIfNeeded();
    }
    else if (response.Action === "create") {
      handleCreateMsg(response);
      generateStartOverlay(overlay, ws, gameId);      
    }
    else if (response.Action === "join") {
      reloadScoreSheet();
      if(mode == "join") {
        generateWaitOverlay(overlay, gameId);
      }
    }
    else if (response.Action === "chat") {
      if(response.Data && response.Data.length < 26 
        && response.Chatter && response.Chatter in playerConIdNameMap) {
        updateMessages(playerConIdNameMap[response.Chatter] + ":  " + response.Data);
      }
    }
    else if (response.Action === "draw") {
      handleDrawMsg(response);
    }
    else if (response.Action === "disconnect") {
      reloadScoreSheet();
    }
    else if (response.Action === "start") {
      hideOverlayElement(overlay);
    }
    else if (response.Action === "roundsame") {
      clearCanvas()
      currentShape = []
      drawing = []
      drawer = (response.Drawer === conId)
      reloadScoreSheet();
      if(drawer) {
        alert("You are drawer!")
      }
    }
    else if (response.Action == "roundchange") {
      clearCanvas()
      currentShape = []
      drawing = []
      drawer = (response.Drawer === conId);
      reloadScoreSheet();
      if(drawer) {
        alert("You are drawer!")
      }
    }
    else if (response.Action === "complete") {
      drawer = false;
      clearCanvas();
      reloadScoreSheet();
      generateGameCompleteOverlay(overlay, playerNameScrHtmlElemMap);
    }
  };
}

function createGameIfNeeded() {
  console.log("connection established");
  if (mode === "create") {
    ws.send(JSON.stringify({
      Action: "create",
      Name: name,
      AvatarId: avatar
    }));
  };
}

async function handleCreateMsg(response) {
  gameId = response.GameId;
  await reloadScoreSheet()
}

function joinGameIfNeeded() {
  if (mode === "join") {
    ws.send(JSON.stringify({
      Action: "join",
      GameId: gameId,
      Name: name,
      AvatarId: avatar
    }));
  };
}

function handleDrawMsg(response) {
  if(drawer || !response.Data) {
    return;
  }
  const img = new Image();
  const elementSrc = `data:image/png;base64,${response.Data}`;
  img.src = elementSrc;
  img.onload = () => {
    clearCanvas();
    canvasContext.drawImage(img, 0, 0, 
      canvasWidth*canvasTopLeft.scaleX, 
      canvasHeight*canvasTopLeft.scaleY);
  }
}

function updateCanvasCordinates() {
  canvasRect = gameCanvas.getBoundingClientRect();
  canvasTopLeft.left = canvasRect.left;
  canvasTopLeft.top = canvasRect.top;
  canvasTopLeft.scaleX = gameCanvas.width / canvasRect.width;
  canvasTopLeft.scaleY = gameCanvas.height / canvasRect.height;
}

function updateDrawing(e) {
  let x1 = e.x - canvasTopLeft.left;
  x1 *= canvasTopLeft.scaleX;
  let y1 = e.y - canvasTopLeft.top;
  y1 *= canvasTopLeft.scaleY;
  if (currentShape.length <= 0) {
    currentShape.push([x1, y1]);
    drawing.push([currentShape, drawItem]);
    return;
  }
  undoWindow = [];
  drawing.pop();
  if (drawItem === "pencil" || drawItem === "brush" || drawItem === "eraser") {
    currentShape.push([x1, y1]);
    drawing.push([currentShape, drawItem]);
    return;
  }
  if (drawItem === "rect" || drawItem === "line") {
    currentShape = currentShape.slice(0, 1);
    currentShape.push([x1, y1]);
    drawing.push([currentShape, drawItem]);
    return;
  }
}

function performUndo() {
  if (drawing.length === 0) {
    return;
  }
  undoWindow.push(drawing.pop());
  updateCanvas();
}

function performRedo() {
  if (undoWindow.length === 0) {
    return;
  }
  drawing.push(undoWindow.pop());
  updateCanvas();
}

function updateDrawingAndClear() {
  drawing.push([null, "clear"]);
  updateCanvas();
}

function updateCanvas() {
  clearCanvas();
  for (const element of drawing) {
    if (element[1] === "clear") {
      clearCanvas();
      continue;
    }
    let elementColor = drawColor;
    if (element[1] == "eraser") {
      elementColor = canvaColor;
    }
    draw(element[0], roughCanvas, element[1], elementColor);
  }
}

function clearCanvas() {
  canvasContext.clearRect(0, 0, canvasWidth, canvasHeight);
}

function updateMessages(msg) {
  const listElement = document.createElement("div");
  listElement.className =
    "text-" + theme + "-500" + " rounded-lg p-2 shadow mx-2 mb-2 mt-2 max-w-sm";
  listElement.textContent = msg;
  if(msgList.childElementCount < 150){
    msgList.appendChild(listElement);
  }
  else{
    msgList.removeChild(msgList.firstElementChild);
    msgList.appendChild(listElement);
  }
  msgList.scrollTop = msgList.scrollHeight;
}

async function displayPlayerScoreCard(playerName, avatarId){
  const response = await fetch(`${httpHost}/avatar?key=${avatarId}`);
  let base64Response = '';
  if(!response.ok){
    console.log("Unexpected Errors: ", response);
  }
  else {
    base64Response = await response.text();
  }
  const elementSrc = `data:image/png;base64,${base64Response}`;

  const scrElement = document.createElement("div");
  scrElement.className = indvScoreBaseClass;

  //Add img element
  {
    const imgElement = document.createElement("img");
    imgElement.alt = "NA";
    imgElement.src = elementSrc;
    imgElement.className = imgElementClass;
    scrElement.appendChild(imgElement);
  }
  //Add name and score
  {

    const nameScoreElem = document.createElement("div");

    const nameElem = document.createElement("div");
    nameElem.textContent = getNameForScore(playerName);
    nameElem.className = nameElemClass;

    const scoreElem = document.createElement("div");
    scoreElem.textContent = 'score:' + '0';
    scoreElem.className = scoreElemClass;

    nameScoreElem.appendChild(nameElem);
    nameScoreElem.appendChild(scoreElem);

    scrElement.appendChild(nameScoreElem);

  }

  playerNameScrHtmlElemMap[playerName] = scrElement;
  scoreSheet.appendChild(scrElement);

}

function getNameForScore(str){
  return str.slice(0,10);
}

function copyGameIdToClipboard(){
  
  const tempElem = document.createElement("textarea");
  tempElem.value = gameId;
  document.body.appendChild(tempElem);
  tempElem.select();
  tempElem.setSelectionRange(0, 99999); // For mobile devices
  document.execCommand('copy');
  document.body.removeChild(tempElem);
  
  let removedClasses = [];
  gameIdDisplay.classList.forEach( classname => {
    if(classname.startsWith("hover:") || (classname.startsWith("bg-"))){
      gameIdDisplay.classList.remove(classname);
      removedClasses.push(classname);
    }
  });

  gameIdDisplay.textContent = "Copied " + gameId;
  gameIdDisplay.classList.add("bg-green-100");
  setTimeout(() => {
    gameIdDisplay.classList.remove("bg-green-100");
    removedClasses.forEach((className) => {
      gameIdDisplay.classList.add(className);
    })
    gameIdDisplay.textContent = "Click To Copy GameId";
  }, 1000)
}

async function reloadScoreSheet() {
  let toRemove = [];
  let toAdd = [];
  const newGameState = await reloadGameStateFromServer(gameId);
  for(const conId of newGameState.ActivePlayers){
    if(gameState?.ActivePlayers.includes(conId)) {
      updateScore(conId, newGameState)
      continue;
    }
    toAdd.push(conId);
  }
  if(newGameState.InactivePlayers?.length > 0) {
  for(const conId of newGameState.InactivePlayers){
      if(gameState?.InactivePlayers?.length > 0 && 
          gameState?.InactivePlayers.includes(conId)
        )
        continue;
      toRemove.push(conId);
    }
  }
  gameState = newGameState;
  if(toAdd?.length >0){
    for(const conId of toAdd){
      const currPlayer = await loadPlayerByConnectionId(conId);
      playerConIdNameMap[conId] = currPlayer.Name
      displayPlayerScoreCard(currPlayer.Name, currPlayer.AvatarId);
    }
  }

  if(toRemove?.length > 0){
    for(const conId of toRemove){
      if(!(conId in playerConIdNameMap))
        continue
      if(!(playerConIdNameMap[conId]) in playerNameScrHtmlElemMap)
        continue
      scoreSheet.removeChild(playerNameScrHtmlElemMap[playerConIdNameMap[conId]])
    }
  }
}

function updateScore(conId, gs={}) {
    if (!(conId in playerConIdNameMap) || !playerConIdNameMap[conId] in playerNameScrHtmlElemMap) {
          return
        }
    let targetElem = playerNameScrHtmlElemMap[playerConIdNameMap[conId]]
    let scr = targetElem.querySelector(".screlem")
    let eScr = 0
    if (conId in gs.PlayerScoreMap) {
      eScr = gs.PlayerScoreMap[conId]
    }
    scr.textContent = "score: " + eScr;
}

function broadcastDrawing() {
  if(!drawer) {
    return;
  }
  ws.send(JSON.stringify({
    Action: "draw",
    GameId: gameId,
    Data: getDrawingData()
  }))
}

function getDrawingData(){
  const base64Data = gameCanvas.toDataURL('image/png').split(',')[1];
  return base64Data;
}

sendDrawUpdates();

function sendDrawUpdates() {
  broadcastDrawing();
  setTimeout(sendDrawUpdates, 1500);
}

function updateTheme() {
  //loadTheme(document.body, theme)
  loadThemeRecursive(gameCanvas, theme);
  loadThemeRecursive(chatButton, theme);
  loadThemeRecursive(chatInput, theme);
  loadThemeRecursive(gameIdDisplay, theme);
  loadThemeRecursive(titleElement, theme);
  loadThemeRecursive(scoreSheet, theme);
}
