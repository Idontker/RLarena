const renderTable = (player) => {
    if (player) {
        // Populate player info
        document.getElementById("player-name").textContent = player.name;
        document.getElementById(
            "player-elo"
        ).textContent = `Current ELO: ${player.current_elo}`;

        // Populate game history table
        const tbody = document.querySelector("#gameHistoryTable tbody");
        const reversedHistory = [...player.game_history].reverse();

        reversedHistory.forEach((game, index) => {
            const eloDiff =
                index === reversedHistory.length - 1
                    ? "+0"
                    : game.elo - reversedHistory[index + 1].elo;
            const eloDiffSign = eloDiff > 0 ? "+" : "";

            const result = game.win ? "W" : game.draw ? "D" : "L";

            const row = `
                <tr>
                    <td>${game.id}</td>
                    <td>${result}</td>
                    <td>${eloDiffSign}${eloDiff}</td>
                    <td>${game.elo}</td>
                </tr>
            `;

            tbody.insertAdjacentHTML("beforeend", row);
        });
    } else {
        document.querySelector(".container").innerHTML =
            "<h2>Player not found</h2>";
    }
};

const onLoad = () => {
    const urlParams = new URLSearchParams(window.location.search);
    const playerId = parseInt(urlParams.get("id"), 10);

    fetch("/users")
        .then((response) => response.json())
        .then((users) => {
            const player = users.find((user) => user.id === playerId);
            renderTable(player);
        });
};
document.addEventListener("DOMContentLoaded", onLoad);
