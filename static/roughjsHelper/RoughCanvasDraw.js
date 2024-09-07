const LINE = "line"
const RECT = "rect"
const PENCIL = "pencil"
const BRUSH = "brush"
const ERASER = "eraser"
const PEN = "pen"

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
        case PEN:
            drawCollectionOfPoints(currentShape, roughCanvas, strokeColor, 5);
        default:
            break;
    }
}

function drawCurve(currentShape, roughCanvas, strokeColor){
    drawCollectionOfPoints(currentShape, roughCanvas, strokeColor, 2);
}

function drawThickCurve(currentShape, roughCanvas, strokeColor){
    drawCollectionOfPoints(currentShape, roughCanvas, strokeColor, 16);
}

function drawEraser(currentShape, roughCanvas, strokeColor){
    drawCollectionOfPoints(currentShape, roughCanvas, strokeColor, 32);
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

function drawCollectionOfPoints(currentShape, roughCanvas, strokeColor, strokeWidth){
    roughCanvas.curve(currentShape, {
                stroke: strokeColor, strokeWidth: strokeWidth, roughness:0
            })
}

export{draw}