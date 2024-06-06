document.addEventListener("DOMContentLoaded", () => {
  const resultsDiv = document.getElementById("results");

  // Create a new WebSocket connection
  const socket = new WebSocket("ws://localhost:8082/ws");

  socket.onopen = () => {
    console.log("WebSocket connection established");

    // Fetch initial vote data via REST API
    fetch('http://localhost:8082/votes')
      .then(response => response.json())
      .then(data => {
        displayResults(data);
      })
      .catch(error => console.error('Error fetching initial data:', error));
  };

  socket.onmessage = (event) => {
    const voteEvent = JSON.parse(event.data);
    console.log("Received vote update:", voteEvent);
    updateResults(voteEvent);
  };

  socket.onclose = () => {
    console.log("WebSocket connection closed");
  };

  socket.onerror = (error) => {
    console.error("WebSocket error:", error);
  };

  function displayResults(data) {
    console.log(data);
    resultsDiv.innerHTML = '';
    for (const { candidateId, count } of data) {
      const candidateDiv = document.createElement('div');
      candidateDiv.className = 'candidate';
      candidateDiv.id = `candidate-${candidateId}`;
      candidateDiv.textContent = `Candidate ${candidateId}: ${count} votes`;
      resultsDiv.appendChild(candidateDiv);
    }
  }

  function updateResults(voteEvent) {
    const candidateDiv = document.getElementById(`candidate-${voteEvent.candidateId}`);
    if (candidateDiv) {
      candidateDiv.textContent = `Candidate ${voteEvent.candidateId}: ${voteEvent.count} votes`;
    } else {
      const newCandidateDiv = document.createElement('div');
      newCandidateDiv.className = 'candidate';
      newCandidateDiv.id = `candidate-${voteEvent.candidateId}`;
      newCandidateDiv.textContent = `Candidate ${voteEvent.candidateId}: ${voteEvent.count} votes`;
      resultsDiv.appendChild(newCandidateDiv);
    }
  }
});
