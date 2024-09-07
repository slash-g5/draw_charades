import rough from "https://cdn.jsdelivr.net/npm/roughjs@4.3.1/bundled/rough.esm.js";
import { draw } from "./roughjsHelper/RoughCanvasDraw.js";

const host = "192.168.0.126"
const wsHost = "ws://" + host + ":8080/ws";
const httpHost = "http://" + host + ":8081" 

const ws = new WebSocket(wsHost);
const urlParams = new URLSearchParams(window.location.search);
const name = urlParams.get("name");
const mode = urlParams.get("mode");
let gameId = urlParams.get("gameId");

const imgElementClass = "h-16 lg:h-24 w:4d";
const nameElemClass = "text-xl font-bold p-2";
const scoreElemClass = "px-2 pb-2 font-bold";
const indvScoreBaseClass = "flex flex-row border-2 border-purple-400 bg-purple-50 rounded-lg p-2 max-h-20 lg:max-h-28 truncate w-52 lg:w-80";

if(mode == null || !(mode === "create" || mode === "join") ||
   name == null || name === "" || 
   (mode === "join" && (gameId == null || gameId === ""))){
  window.location.href = "http://" + host + ":8081/welcome"
}

const avatar = urlParams.get("avatar");
let fellowGamersMap = {}

//gameId display
let gameIdDisplay = document.getElementById("game_id_display");

//Chat window
const msgList = document.getElementById("msg_list");

//Score List
const scoreSheet = document.getElementById("score_sheet");
let score = []

//Canvas related
let drawer = true;
let isDrawing = false;
let drawItem = "pencil";
let drawColor = "#8B5CF6";
let canvaColor = "#F3E8FF";
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

//websocket message Actions
{
  ws.onmessage = (message) => {
    console.log(message.data);
    const response = JSON.parse(message.data);
    if (response.Action === "connect"){
      createGameIfNeeded();
      joinGameIfNeeded();
    }
    else if (response.Action === "create"){
      handleCreateMsg(response);
    }
  };
}

function createGameIfNeeded() {
  console.log("connection established");
  if (mode === "create") {
    ws.send(JSON.stringify({
      Action: "create"
    }));
  };
}

function handleCreateMsg(response) {
  gameId = response.GameId;
  displayGameId();
}

function joinGameIfNeeded() {
  displayGameId();
}

function displayGameId(){
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
    "text-violet-500 rounded-lg p-2 shadow mx-2 mb-2 mt-2 max-w-sm";
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

async function displaySelfScoreCard(){
  const response = await fetch(`${httpHost}/avatar?key=${avatar}`);
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
    nameElem.textContent = getNameForScore(name);
    nameElem.className = nameElemClass;

    const scoreElem = document.createElement("div");
    scoreElem.textContent = 'score:' + '0';
    scoreElem.className = scoreElemClass;

    nameScoreElem.appendChild(nameElem);
    nameScoreElem.appendChild(scoreElem);

    scrElement.appendChild(nameScoreElem);

  }

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
    if(classname.startsWith("hover:")){
      gameIdDisplay.classList.remove(classname);
      removedClasses.push(classname);
    }
  });

  gameIdDisplay.classList.add("bg-green-100");
  setTimeout(() => {
    gameIdDisplay.classList.remove("bg-green-100");
    removedClasses.forEach((className) => {
      gameIdDisplay.classList.add(className);
    })
  }, 1000)
}

displaySelfScoreCard();