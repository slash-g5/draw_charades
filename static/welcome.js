import { draw } from "./roughjsHelper/RoughCanvasDraw.js";
import rough from "https://cdn.jsdelivr.net/npm/roughjs@4.3.1/bundled/rough.esm.js";

const host = "http://192.168.0.126";
const cButton = document.getElementById("create_button");
const jButton = document.getElementById("join_button");
const name = document.getElementById("name");
const gameId = document.getElementById("game_id");

const avatarCanvas = document.getElementById("avatar_canvas");
const clearTool = document.getElementById("clear_tool");
const roughCanvas = rough.canvas(avatarCanvas);
const canvasContext = avatarCanvas.getContext("2d");

const drawItem = "pencil";
let isDrawing = false;
let avatar = []
let currentShape = []
let drawColor = "#8B5CF6";

let canvasRect = avatarCanvas.getBoundingClientRect();
const canvasWidth = canvasRect.width;
const canvasHeight = canvasRect.height;
const canvasTopLeft = {
  top: 0,
  left: 0,
  scaleX: 0,
  scaleY: 0,
};

cButton.addEventListener("click", () => {
  if (!name.value) {
    alert("Please Enter Name");
    return;
  }
  getAvatarImageId().then(imageId => {
    window.location.href = host + ":8081/create?name=" + name.value + "&avatar="+imageId;
  }).catch(e => console.error("Error:", e));
});

jButton.addEventListener("click", () => {
  if (!name.value) {
    alert("Please Enter Name");
    return;
  }
  if (!gameId.value) {
    alert("Please Enter GameId");
    return;
  }
  getAvatarImageId().then(imageId => {
      window.location.href = host + ":8081?mode=join&name=" + name.value +
    "&gameId=" + gameId.value + "&avatar="+imageId;
  }).catch(e => console.error("Error:", e)) 
});

//avatarCanvas Event Listeners
{
  avatarCanvas.addEventListener("mousemove", (e) => {
    updateCanvasCordinates();
    if (!isDrawing) {
      return;
    }
    updateAvatar(e);
    updateCanvas();
  });

  avatarCanvas.addEventListener("touchmove", (eTouch) => {
    updateCanvasCordinates();
    eTouch.preventDefault();
    let e = {
      "x": eTouch.touches[0].clientX,
      "y": eTouch.touches[0].clientY,
    };
    updateAvatar(e);
    updateCanvas();
  });

  avatarCanvas.addEventListener("mousedown", (_) => {
    isDrawing = true;
  });

  avatarCanvas.addEventListener("mouseup", (_) => {
    isDrawing = false;
    currentShape = [];
  });

  avatarCanvas.addEventListener("mouseleave", (_) => {
    isDrawing = false;
    currentShape = [];
  });

  avatarCanvas.addEventListener("touchend", (_) => {
    isDrawing = false;
    currentShape = [];
  });
}

clearTool.addEventListener("click", () => {
  updateAvatarAndClear();
});

function updateCanvasCordinates() {
  canvasRect = avatarCanvas.getBoundingClientRect();
  canvasTopLeft.left = canvasRect.left;
  canvasTopLeft.top = canvasRect.top;
  canvasTopLeft.scaleX = avatarCanvas.width / canvasRect.width;
  canvasTopLeft.scaleY = avatarCanvas.height / canvasRect.height;
}

function updateAvatar(e) {
  let x1 = e.x - canvasTopLeft.left;
  x1 *= canvasTopLeft.scaleX;
  let y1 = e.y - canvasTopLeft.top;
  y1 *= canvasTopLeft.scaleY;
  if (currentShape.length <= 0) {
    currentShape.push([x1, y1]);
    avatar.push([currentShape, drawItem]);
    return;
  }
  avatar.pop(); 
  currentShape.push([x1, y1]);
  avatar.push([currentShape, drawItem]);
  return;
}

function updateAvatarAndClear() {
  avatar = [];
  updateCanvas();
}

function updateCanvas() {
  clearCanvas();
  for (const element of avatar) {
    draw(element[0], roughCanvas, element[1], drawColor);
  }
}

function clearCanvas() {
  canvasContext.clearRect(0, 0, canvasWidth*canvasTopLeft.scaleX, canvasHeight*canvasTopLeft.scaleY);
}

function getAvatarString(){
  const base64Avatar = avatarCanvas.toDataURL('image/png').split(',')[1];
  return base64Avatar;
}

function getAvatarImageId(){
  const requestUrl = host + ":8081/avatar";
  return fetch(requestUrl, {
    method : "POST",
    headers : {
      "Content-Type" : "text/plain"
    },
    body : getAvatarString()
  })
  .then(data => {
    if(!data.ok){
      throw new Error("Unexpected responseData")
    }
    return data.text();
  })
  .catch(err => {
    console.error('Error:', err)
    return null
  })
}
