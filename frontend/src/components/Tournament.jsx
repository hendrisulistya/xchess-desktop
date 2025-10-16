import React, { useEffect, useState } from "react";
import {
  ListPlayers,
  InitTournamentWithPlayerIDs,
  NextRound,
  GetCurrentRound,
  RecordResult,
  GetPlayers,
} from "../../wailsjs/go/main/App";

function Tournament() {
  const [players, setPlayers] = useState([]);
  const [selectedIDs, setSelectedIDs] = useState([]);
  const [currentRound, setCurrentRound] = useState(null);
  const [standings, setStandings] = useState([]);
  // Add status and ID->Name map
  const [status, setStatus] = useState("");
  const [idToName, setIdToName] = useState({});

  useEffect(() => {
    ListPlayers()
      .then((list) => {
        setPlayers(list || []);
        // Build quick map for display using lowercase keys from backend JSON
        const map = {};
        (list || []).forEach((p) => (map[p.id] = p.name));
        setIdToName(map);
      })
      .catch((err) => console.error("ListPlayers error:", err));
  }, []);

  const toggleSelect = (id) => {
    setSelectedIDs((prev) =>
      prev.includes(id) ? prev.filter((x) => x !== id) : [...prev, id]
    );
  };

  const initTournament = async () => {
    try {
      setStatus("Initializing tournament...");
      const ok = await InitTournamentWithPlayerIDs("Turnamen 1", selectedIDs);
      if (!ok) {
        setStatus("Failed to initialize. Select players first.");
        return;
      }
      setStatus("Tournament initialized. Generating first round...");
      // Immediately generate Round 1 so matches appear
      const roundOk = await NextRound();
      if (!roundOk) {
        setStatus("Failed to generate first round.");
        return;
      }
      setStatus("Round 1 generated.");
      await refreshRoundAndStandings();
    } catch (e) {
      console.error("InitTournamentWithPlayerIDs error:", e);
      setStatus("Error initializing tournament.");
    }
  };

  const nextRound = async () => {
    try {
      setStatus("Generating next round...");
      const ok = await NextRound();
      if (!ok) {
        setStatus("Failed to generate next round.");
        return;
      }
      setStatus("Next round generated.");
      await refreshRoundAndStandings();
    } catch (e) {
      console.error("NextRound error:", e);
      setStatus("Error generating next round.");
    }
  };

  const recordResult = async (tableNumber, result) => {
    try {
      setStatus(`Recording result for table ${tableNumber}...`);
      const ok = await RecordResult(tableNumber, result);
      if (!ok) {
        setStatus("Failed to record result.");
        return;
      }
      setStatus("Result recorded.");
      await refreshRoundAndStandings();
    } catch (e) {
      console.error("RecordResult error:", e);
      setStatus("Error recording result.");
    }
  };

  const refreshRoundAndStandings = async () => {
    try {
      const round = await GetCurrentRound();
      setCurrentRound(round);
      const ps = await GetPlayers();
      // Keep name map fresh from tournament players
      const map = {};
      (ps || []).forEach((p) => (map[p.id] = p.name));
      setIdToName(map);
      ps.sort((a, b) => {
        if (a.score !== b.score) return b.score - a.score;
        if (a.buchholz !== b.buchholz) return b.buchholz - a.buchholz;
        if (a.rating !== b.rating) return b.rating - a.rating;
        return a.name.localeCompare(b.name);
      });
      setStandings(ps);
    } catch (e) {
      console.error("Refresh error:", e);
      setStatus("Error refreshing round/standings.");
    }
  };

  return (
    <div>
      {/* Status helper */}
      {status && <div style={{ marginBottom: 12 }}>{status}</div>}

      <h2>Peserta</h2>
      <ul>
        {players.map((p) => (
          <li key={p.id}>
            <label>
              <input
                type="checkbox"
                checked={selectedIDs.includes(p.id)}
                onChange={() => toggleSelect(p.id)}
              />
              {p.name} (Rating: {p.rating})
            </label>
          </li>
        ))}
      </ul>
      <button onClick={initTournament} disabled={selectedIDs.length < 1}>
        Initialize Tournament
      </button>

      <hr />

      <h2>Current Round</h2>
      {!currentRound || !currentRound.matches || currentRound.matches.length === 0 ? (
        <p>No matches yet. Initialize tournament, then Round 1 will be generated automatically.</p>
      ) : (
        <div>
          <p>Round #{currentRound.round_number}</p>
          <table>
            <thead>
              <tr>
                <th>Table</th>
                <th>Player A</th>
                <th>Player B</th>
                <th>White</th>
                <th>Black</th>
                <th>Result</th>
              </tr>
            </thead>
            <tbody>
              {currentRound.matches.map((m) => (
                <tr key={m.table_number}>
                  <td>{m.table_number}</td>
                  {/* Show names instead of IDs using lowercase keys */}
                  <td>{idToName[m.player_a_id] || m.player_a_id}</td>
                  <td>{idToName[m.player_b_id] || m.player_b_id}</td>
                  <td>{idToName[m.white_id] || m.white_id}</td>
                  <td>{idToName[m.black_id] || m.black_id}</td>
                  <td>
                    {m.player_b_id === "BYE" ? (
                      <button onClick={() => recordResult(m.table_number, "BYE_A")}>Apply Bye</button>
                    ) : (
                      <>
                        <button onClick={() => recordResult(m.table_number, "A_WIN")}>A Win</button>
                        <button onClick={() => recordResult(m.table_number, "B_WIN")}>B Win</button>
                        <button onClick={() => recordResult(m.table_number, "DRAW")}>Draw</button>
                      </>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          <button onClick={nextRound}>Generate Next Round</button>
        </div>
      )}

      <hr />

      <h2>Standings</h2>
      <ol>
        {standings.map((p) => (
          <li key={p.id}>
            {p.name} — {p.score} pts (Buchholz: {p.buchholz}, PS: {p.progressive_score ? p.progressive_score.toFixed(1) : '0.0'})
          </li>
        ))}
      </ol>
    </div>
  );
}

export default Tournament;