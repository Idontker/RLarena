const renderLeaderboard = (users) => {
    console.log(users);
    // Sort users by current ELO
    users.sort((a, b) => b.current_elo - a.current_elo);

    // Generate table rows
    const tbody = document.querySelector("#playerTable tbody");
    tbody.innerHTML = "";

    users.forEach((player) => {
        const totalGames = player.game_history.length;
        const wins = player.game_history.filter((game) => game.win).length;
        const draws = player.game_history.filter((game) => game.draw).length;
        const losses = player.game_history.filter((game) => game.loss).length;
        const winPercentage = ((wins / totalGames) * 100).toFixed(1);
        const drawPercentage = ((draws / totalGames) * 100).toFixed(1);
        const lossPercentage = ((losses / totalGames) * 100).toFixed(1);

        const lastFiveGames = player.game_history
            .slice(-5)
            .map((game) => {
                if (game.win) return "W";
                if (game.draw) return "D";
                if (game.loss) return "L";
            })
            .join(", ");

        // TODO: can be attacked with a stored XSS attack on player.name ?
        const row = `
             <tr>
                 <td>${player.id}</td>
                 <td>${player.current_elo}</td>
                 <td>${player.name}</td>
                 <td>${lastFiveGames}</td>
                 <td>${winPercentage}% / ${drawPercentage}% / ${lossPercentage}%</td>
                 <td>${totalGames}</td>
                 <td>
                 <a href="/player?id=${player.id}">
                    <span class="info-icon" title="More info about ${player.name}">ℹ️</span>
                    View
                 </a>
                 </td>
             </tr>
         `;

        tbody.insertAdjacentHTML("beforeend", row);
    });
};

const onLoad = () => {
    fetch("/users")
        .then((response) => response.json())
        .then((users) => {
            renderLeaderboard(users);
        });
};
document.addEventListener("DOMContentLoaded", () => {
    // // const users = [
    //     {
    //         id: 1,
    //         name: "rado1",
    //         current_elo: 1080,
    //         game_history: [
    //             { id: 9, win: true, draw: false, loss: false, elo: 1032 },
    //             { id: 9, win: false, draw: false, loss: true, elo: 1000 },
    //             { id: 5, win: false, draw: true, loss: false, elo: 1016 },
    //             { id: 6, win: true, draw: false, loss: false, elo: 1048 },
    //             { id: 11, win: true, draw: false, loss: false, elo: 1080 },
    //             { id: 11, win: false, draw: false, loss: true, elo: 1048 },
    //             { id: 3, win: false, draw: false, loss: true, elo: 1048 },
    //             { id: 4, win: false, draw: false, loss: true, elo: 1048 },
    //             { id: 10, win: false, draw: false, loss: true, elo: 1048 },
    //             { id: 10, win: true, draw: false, loss: false, elo: 1176 },
    //             { id: 2, win: true, draw: false, loss: false, elo: 1208 },
    //             { id: 7, win: true, draw: false, loss: false, elo: 1240 },
    //             { id: 8, win: false, draw: false, loss: true, elo: 1240 },
    //             { id: 8, win: true, draw: false, loss: false, elo: 1080 },
    //         ],
    //     },
    //     {
    //         id: 2,
    //         name: "primo1",
    //         current_elo: 1240,
    //         game_history: [
    //             { id: 9, win: true, draw: false, loss: false, elo: 1032 },
    //             { id: 9, win: false, draw: false, loss: true, elo: 1000 },
    //             { id: 5, win: false, draw: true, loss: false, elo: 1016 },
    //             { id: 6, win: true, draw: false, loss: false, elo: 1048 },
    //             { id: 11, win: true, draw: false, loss: false, elo: 1080 },
    //             { id: 11, win: false, draw: false, loss: true, elo: 1048 },
    //             { id: 3, win: false, draw: false, loss: true, elo: 1048 },
    //             { id: 4, win: false, draw: false, loss: true, elo: 1048 },
    //             { id: 10, win: false, draw: false, loss: true, elo: 1048 },
    //             { id: 10, win: true, draw: false, loss: false, elo: 1176 },
    //             { id: 2, win: true, draw: false, loss: false, elo: 1208 },
    //             { id: 7, win: true, draw: false, loss: false, elo: 1240 },
    //             { id: 8, win: false, draw: false, loss: true, elo: 1240 },
    //         ],
    //     },
    // ];
    onLoad();
});
