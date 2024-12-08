document.addEventListener("DOMContentLoaded", () => {
    const gameTableBody = document.querySelector("#gameTable tbody");

    // Get the page number from the query parameter
    const params = new URLSearchParams(window.location.search);
    const page = params.get("page");
    const pageNumber = Number(page);

    // Check if the page number is valid
    const apiUrl =
        pageNumber && !isNaN(pageNumber)
            ? `/games/all?page=${pageNumber}`
            : "/games/all?page=1";

    // Fetch games from the API with the page parameter if valid
    fetch(apiUrl)
        .then((response) => {
            if (!response.ok) {
                throw new Error("Network response was not ok");
            }
            return response.json();
        })
        .then((games) => {
            // Populate the table
            games.forEach((game) => {
                console.log(game);
                const row = document.createElement("tr");

                // Create cells for each field

                const infoCell = document.createElement("td");
                infoCell.innerHTML = `<a href="/game?id=${game.id}">
                    <span class="info-icon" title="More info about game ${game.id}">ℹ️</span>
                    ${game.id}
                </a>`;

                const player1IdCell = document.createElement("td");
                player1IdCell.textContent = game.player1_id;

                const player2IdCell = document.createElement("td");
                player2IdCell.textContent = game.player2_id;

                const outcomeCell = document.createElement("td");
                outcomeCell.textContent = getOutcomeText(game.outcome);

                // Append cells to the row
                row.appendChild(infoCell);
                row.appendChild(player1IdCell);
                row.appendChild(player2IdCell);
                row.appendChild(outcomeCell);

                // Append the row to the table body
                gameTableBody.appendChild(row);
            });
        })
        .catch((error) => {
            console.error("Error fetching game data:", error);
        });

    // Helper function to convert outcome codes to text
    function getOutcomeText(outcome) {
        switch (outcome) {
            case -1:
                return "Draw";
            case 0:
                return "Ongoing";
            case 1:
                return "Player 1 Wins";
            case 2:
                return "Player 2 Wins";
            default:
                return "Unknown";
        }
    }
});
