import { httpHost } from "../config/uiConfig.js";

async function reloadGameStateFromServer(gameId) {
    const response = await fetch(`${httpHost}/game?key=${gameId}`);
    const jsonRes = await response.json();
    return jsonRes;
}

async function loadPlayerByConnectionId(conId) {
    const response = await fetch(`${httpHost}/player?key=${conId}`);
    const jsonRes = await response.json();
    return jsonRes;
}

async function loadAvatarImageById(avatarId) {
    const response = await fetch(`${httpHost}/avatar?key=${avatarId}`);
    let base64Response = '';
    if(!response.ok){
        console.log("Unexpected Errors: ", response);
        return base64Response
    }
    else {
        base64Response = await response.text();
    }
    const elementSrc = `data:image/png;base64,${base64Response}`;
    return elementSrc
}

export {
            reloadGameStateFromServer, 
            loadPlayerByConnectionId,
            loadAvatarImageById
        }