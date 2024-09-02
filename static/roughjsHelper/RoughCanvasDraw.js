const LINE = "line"
const RECT = "rect"
const PENCIL = "pencil"
const BRUSH = "brush"
const ERASER = "eraser"

function draw(currentShape, roughCanvas, type, strokeColor="red"){
    switch (type) {
        case LINE:
            drawLine(currentShape, roughCanvas, strokeColor);
            break;
        case RECT:
            drawRect(currentShape, roughCanvas, strokeColor);
            break;
        case PENCIL:
            drawCurve(currentShape, roughCanvas, strokeColor);
            break;
        case BRUSH:
            drawThickCurve(currentShape, roughCanvas, strokeColor);
            break;
        case ERASER:
            drawEraser(currentShape, roughCanvas, strokeColor);
        default:
            break;
    }
}

function drawCurve(currentShape, roughCanvas, strokeColor){
    roughCanvas.curve(currentShape, {
  stroke: strokeColor, strokeWidth: 2, roughness:0
})
}

function drawThickCurve(currentShape, roughCanvas, strokeColor){
    roughCanvas.curve(currentShape, {
  stroke: strokeColor, strokeWidth: 16, roughness:0
})
}

function drawEraser(currentShape, roughCanvas, strokeColor){
    roughCanvas.curve(currentShape, {
  stroke: strokeColor, strokeWidth: 32, roughness:0
})
}

function drawRect(currentShape, roughCanvas, strokeColor) {

    if(currentShape.length <2){
        return;
    }

    roughCanvas.rectangle(
        currentShape[0][0],    
        currentShape[0][1],
        currentShape[1][0]-currentShape[0][0],
        currentShape[1][1]-currentShape[0][1],
        {
            stroke: strokeColor,
            strokeWidth: 2,
            roughness:0
        });
}

function drawLine(currentShape, roughCanvas, strokeColor){

    if(currentShape.length < 2){
        return;
    }

    roughCanvas.line(currentShape[0][0],
        currentShape[0][1],
        currentShape[1][0],
        currentShape[1][1],
        {
            stroke: strokeColor,
            strokeWidth: 2,
            roughness: 0
        });
}

export{draw}