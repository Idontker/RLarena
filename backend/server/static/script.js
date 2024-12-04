let originalGameState = null;

// Fetch game state by ID
async function fetchGameState(id) {
    const response = await fetch(`/game/${id}/state`);
    if (response.ok) {
        return await response.json();
    } else {
        alert(`Failed to load game state for ID ${id}`);
        return null;
    }
}

// Render the board
function renderBoard(board, rows, cols) {
    const boardDiv = document.getElementById("board");
    boardDiv.innerHTML = "";

    for (let row = rows - 1; row >= 0; row--) {
        const rowDiv = document.createElement("div");
        rowDiv.classList.add("row");

        for (let col = 0; col < cols; col++) {
            const cell = document.createElement("div");
            cell.classList.add("cell");

            // Determine the checker pattern
            if ((row + col) % 2 === 0) {
                cell.classList.add("white");
            } else {
                cell.classList.add("black");
            }

            // Add game piece if present
            if (board[row][col] === 1) {
                cell.innerHTML = '<span class="player1">1</span>';
            } else if (board[row][col] === 2) {
                cell.innerHTML = '<span class="player2">2</span>';
            }

            rowDiv.appendChild(cell);
        }

        boardDiv.appendChild(rowDiv);
    }
}

// Render move history
function renderHistory(history) {
    const player1HistoryDiv = document.getElementById("player1-history");
    const player2HistoryDiv = document.getElementById("player2-history");

    player1HistoryDiv.innerHTML = "";
    player2HistoryDiv.innerHTML = "";

    history.forEach((turn, index) => {
        const button = document.createElement("button");
        button.classList.add("move");
        const div1 = document.createElement("div");
        const div2 = document.createElement("div");
        div1.textContent = `Move ${index + 1}`;
        div2.textContent = `(${turn.sourceRow}, ${turn.sourceCol}) → (${turn.destRow}, ${turn.destCol})`;
        button.appendChild(div1);
        button.appendChild(div2);

        button.onclick = () => {
            // Reconstruct the game state up to this turn
            reconstructGameState(history.slice(0, index + 1));
        };

        if (turn.player === 1) {
            player1HistoryDiv.appendChild(button);
        } else {
            player2HistoryDiv.appendChild(button);
        }
    });
}

function renderSlider(history) {
    const slider = document.getElementById("history-range");
    slider.min = 0;
    slider.max = history.length;
    slider.value = history.length;
    slider.step = 1;

    document.getElementById("currentMoveSpan").textContent = history.length;

    slider.oninput = () => {
        const value = parseInt(slider.value);
        reconstructGameState(history.slice(0, value));
    };
}

// Reconstruct the game state based on a subset of history
function reconstructGameState(partialHistory) {
    // Clone the original state
    const newState = JSON.parse(JSON.stringify(originalGameState));
    newState.board = newState.board.map((row) => row.map(() => 0));
    for (let col = 0; col < newState.cols; col++) {
        newState.board[0][col] = 1;
        newState.board[newState.rows - 1][col] = 2;
    }

    // Apply each turn to the board
    partialHistory.forEach((turn) => {
        newState.board[turn.sourceRow][turn.sourceCol] = 0;
        newState.board[turn.destRow][turn.destCol] = turn.player;
    });

    document.getElementById("currentMoveSpan").textContent =
        partialHistory.length;

    renderBoard(newState.board, newState.rows, newState.cols);
}

// Load game state and render everything
async function loadGameState() {
    const urlParams = new URLSearchParams(window.location.search);
    const gameId = urlParams.get("id");

    if (gameId == "") {
        return;
    }

    if (!gameId) {
        alert("Please enter a valid game ID.");
        return;
    }

    const game = await fetchGameState(gameId);
    const gameState = game["game_state"];

    if (gameState) {
        originalGameState = gameState;
        renderBoard(gameState.board, gameState.rows, gameState.cols);
        renderHistory(gameState.history);
        renderSlider(gameState.history);
    }
}

// Attach event listeners
const redirectToGame = (event) => {
    event.preventDefault();
    const gameId = document.getElementById("game-selector").value;
    // alert(gameId);
    if (gameId) {
        window.location.href = `?id=${gameId}`;
    } else {
        alert("Please enter a valid game ID.");
    }
};
document
    .getElementById("selectGame")
    .addEventListener("submit", redirectToGame);
document.getElementById("load-game").addEventListener("click", redirectToGame);

// Load the game state on page load
window.addEventListener("load", loadGameState);
