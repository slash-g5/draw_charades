import rough from "https://cdn.jsdelivr.net/npm/roughjs@4.3.1/bundled/rough.esm.js";
import { draw } from "./roughjsHelper/RoughCanvasDraw.js";

const host = "192.168.0.126"
const wsHost = "ws://" + host + ":8080/ws";
const httpHost = "http://" + host + ":8081" 

const ws = new WebSocket(wsHost);
const urlParams = new URLSearchParams(window.location.search);
const name = urlParams.get("name");
const mode = urlParams.get("mode");
const avatar = urlParams.get("avatar");
let fellowGamersMap = {}

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

//websocket message Actions
{
  ws.onmessage = (message) => {
    const response = message.data;
    console.log(response);
  };
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

function updateScore() {
  scoreSheet.innerHTML = '';
  for(const scr of score){
    const scoreElem = document.createElement("div");
    scoreElem.className = 
    "text-violet-500 rounded-lg p-2 shadow mx-2 mb-2 mt-2 max-w-sm";
    scoreElem.textContent = `${scr[0]} ${scr[1]}`
    scoreSheet.appendChild(scoreElem);
  }
}

async function displayAvatar(){
  const response = await fetch(`${httpHost}/avatar?key=${avatar}`);
  if(!response.ok){
    console.error("Unexpected Errors")
  }
  const base64Response = await response.text();
  console.log(base64Response);
  return base64Response;
}
displayAvatar();