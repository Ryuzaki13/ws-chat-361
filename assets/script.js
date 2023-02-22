const messageContainer = document.querySelector("#MessageLog");
const messageInput = document.querySelector("#Message");
const sendButton = document.querySelector("#SendMessage");
const typingInfo = document.querySelector("#TypingInfo");
const game = document.querySelector(".game");

/**
 * Список активных игроков
 * @type {Map<string, HTMLDivElement>}
 */
const playerList = new Map();
/**
 * Для отслеживания пользователей печатающих в чат
 * @type {Map<string, number>}
 */
let typingUsers = new Map();

let keyPressed = {
    ArrowLeft: false,
    ArrowRight: false,
    ArrowUp: false,
    ArrowDown: false,
}

const Users = BuildUserList();
const player = CreatePlayer({radius: 100, color: "#FF0000", name: playerName});

game.append(player);
playerList.set(playerName, player);

const url = "ws://127.0.0.1:8085/connect/" + playerName;
let ws = new WebSocket(url);


const MESSAGE_SEND = 1;
const MESSAGE_TYPING = 2;
const UPDATE_GAME = 3;
const ONLINE_LIST = 4;

ws.onmessage = e => {
    const transfer = JSON.parse(e.data);
    const dataMessage = JSON.parse(transfer.data);

    switch (transfer.type) {
        case MESSAGE_SEND:
            createMessage(dataMessage);
            break;
        case MESSAGE_TYPING:
            typingUsers.set(dataMessage.name, 0);

            updateTyping();

            setTimeout(() => {
                typingUsers.clear();
                typingInfo.textContent = "";
            }, 3000);

            break;
        case UPDATE_GAME:
            UpdateFrame(dataMessage)
            break;
        case ONLINE_LIST:
            Users.UpdateUserStatus(dataMessage);
    }
}
ws.onclose = e => {
    setTimeout(() => {
        ws = new WebSocket(url);
    }, 1000);
}
sendButton.onclick = () => {
    const message = messageInput.value.trim();
    if (message.length > 0) {
        ws.send(JSON.stringify({
            type: 1,
            data: JSON.stringify({
                name: playerName,
                message: message,
            })
        }));
        createMessage({message: message, self: true});
        messageInput.value = "";
    }
}
messageInput.oninput = () => {
    if (messageInput.value.length > 0) {
        ws.send(JSON.stringify({
            type: 2,
            data: JSON.stringify({
                name: playerName,
            })
        }));
    }
}

document.onkeydown = (e) => {
    keyPressed[e.code] = true;
}
document.onkeyup = (e) => {
    keyPressed[e.code] = false;
}

function createMessage(data) {
    const messageDiv = document.createElement("div");
    const userName = document.createElement("span");
    const userMsg = document.createElement("span");
    const userTime = document.createElement("div");

    messageDiv.append(userTime, userName, userMsg);

    if (data.self === true) {
        messageDiv.className = "self";
        userName.textContent = "";
        data.time = new Date().toJSON();
    } else {
        userName.textContent = data.name + ": ";
    }

    userMsg.textContent = data.message;
    userTime.textContent = GetDateTimeFormat(data.time);
    userTime.className = "time";
    userName.className = "name";
    userMsg.className = "message";

    messageContainer.append(messageDiv);
}

function updateTyping() {
    let typing = [];
    for (let [k, v] of typingUsers) {
        typing.push(k);
    }
    if (typingUsers.size > 1) {
        typingInfo.textContent =
            typing.join(", ") + " набирают сообщение";
    } else {
        typingInfo.textContent =
            typing.join(", ") + " набирает сообщение";
    }
}

function UpdateFrame(players) {
    for (let player of players) {
        const p = playerList.get(player.name);
        if (player.isDead === true) {
            if (p) p.remove();
            playerList.delete(player.name);
            continue
        }
        if (p) {
            p.style.left = player.coords.X + "px";
            p.style.top = player.coords.Y + "px";
            p.style.height = player.radius + "px";
            p.style.width = player.radius + "px";
        } else {
            const p1 = CreatePlayer(player);

            game.append(p1);
            playerList.set(player.name, p1);
        }
    }
}

function CreatePlayer(player) {
    const playerElement = document.createElement("div");
    playerElement.className = "player";
    playerElement.style.width = player.radius + "px";
    playerElement.style.height = player.radius + "px";
    playerElement.style.backgroundColor = player.color;
    playerElement.textContent = player.name;
    return playerElement;
}

function SendCoords() {
    let direction = {
        X: 0,
        Y: 0
    }
    if (keyPressed.ArrowLeft === true ^ keyPressed.ArrowRight === true) {
        if (keyPressed.ArrowLeft === true) {
            direction.X = -1;
        } else {
            direction.X = 1;
        }
    }
    if (keyPressed.ArrowUp === true ^ keyPressed.ArrowDown === true) {
        if (keyPressed.ArrowUp === true) {
            direction.Y = -1;
        } else {
            direction.Y = 1;
        }
    }

    ws.send(JSON.stringify({
        type: 3,
        data: JSON.stringify(direction)
    }));
}

function GetDateTimeFormat(timestamp) {
    if (!timestamp || timestamp.length < 20) return "";

    const date = new Date(timestamp);

    return "" + date.getDate() + months[date.getMonth()] + date.getFullYear() +
        " " + date.getHours() + ":" + date.getMinutes() + ":" + date.getSeconds();
}

const months = [
    " января ",
    " февраля ",
    " марта ",
    " апреля ",
    " мая ",
    " июня ",
    " июля ",
    " августа ",
    " сентября ",
    " октября ",
    " нобяря ",
    " декабря "
];

setInterval(SendCoords, 25);

/**
 * Создание объекта обрабатывающий список онлайн-пользователей
 */
function BuildUserList() {
    const userListElement = document.querySelector(".user-list");
    if (!userListElement) return;

    /**
     * @type {{[Name: string]: HTMLDivElement}}
     */
    const userList = {};

    const thisObject = {};

    thisObject.UpdateUserStatus = function (users) {
        clearOnlineStatus();

        if (users && users.length > 0) {
            for (let username of users) {
                const element = userList[username];
                if (element) {
                    element.classList.add("online");
                }
            }
        }
    }

    function clearOnlineStatus() {
        for (let user in userList) {
            userList[user].classList.remove("online");
        }
    }

    function createUserList(users) {
        ClearElement(userListElement);

        if (users && users.length > 0) {
            for (let username of users) {
                const userItem = createUserItem(username);

                userList[username] = userItem;
                userListElement.append(userItem);
            }
        }
    }

    function createUserItem(name) {
        const item = document.createElement("div");
        item.className = "user";
        item.textContent = name;
        return item;
    }

    function getUsers() {
        const xhr = new XMLHttpRequest();
        xhr.open("GET", "/user/list")
        xhr.onload = () => {
            createUserList(JSON.parse(xhr.response));

            ws.send(JSON.stringify({
                type: ONLINE_LIST,
                data: null,
            }));
        }
        xhr.send();
    }

    getUsers();

    return thisObject;
}

function ClearElement(element) {
    if (!element) return;
    while (element.childNodes.length > 0) {
        element.childNodes[0].remove();
    }
}