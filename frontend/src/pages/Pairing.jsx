import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router";
import {
  ListPlayers,
  InitTournamentWithPlayerIDs,
  NextRound,
  GetCurrentRound,
  RecordResult,
  GetPlayers,
  GetTournamentInfo,
} from "../../wailsjs/go/main/App";

function Pairing() {
  const navigate = useNavigate();
  const [players, setPlayers] = useState([]);
  const [selectedIDs, setSelectedIDs] = useState([]);
  const [currentRound, setCurrentRound] = useState(null);
  const [standings, setStandings] = useState([]);
  const [tournamentInfo, setTournamentInfo] = useState(null);
  const [status, setStatus] = useState("");
  const [idToName, setIdToName] = useState({});
  const [activeTab, setActiveTab] = useState("setup");

  useEffect(() => {
    loadInitialData();
  }, []);

  const loadInitialData = async () => {
    try {
      const playerList = await ListPlayers();
      setPlayers(playerList || []);
      const map = {};
      (playerList || []).forEach((p) => (map[p.id] = p.name));
      setIdToName(map);

      const tournamentData = await GetTournamentInfo();
      if (tournamentData) {
        setTournamentInfo(tournamentData);
        setActiveTab("pairing");
        await refreshRoundAndStandings();
      }
    } catch (error) {
      console.error("Error loading initial data:", error);
      setStatus("Error loading data");
    }
  };

  const toggleSelect = (id) => {
    setSelectedIDs((prev) =>
      prev.includes(id) ? prev.filter((x) => x !== id) : [...prev, id]
    );
  };

  const initTournament = async () => {
    if (selectedIDs.length < 2) {
      setStatus("Pilih minimal 2 pemain untuk memulai turnamen");
      return;
    }

    try {
      setStatus("Menginisialisasi turnamen...");
      const ok = await InitTournamentWithPlayerIDs(
        "Turnamen Baru",
        "Turnamen catur dengan sistem Swiss",
        selectedIDs
      );
      if (!ok) {
        setStatus("Gagal menginisialisasi turnamen");
        return;
      }

      setStatus("Turnamen berhasil dibuat. Membuat ronde pertama...");
      const roundOk = await NextRound();
      if (!roundOk) {
        setStatus("Gagal membuat ronde pertama");
        return;
      }

      setStatus("Ronde 1 berhasil dibuat");
      setActiveTab("pairing");
      await refreshRoundAndStandings();
    } catch (error) {
      console.error("Error initializing tournament:", error);
      setStatus("Error saat menginisialisasi turnamen");
    }
  };

  const nextRound = async () => {
    try {
      setStatus("Membuat ronde berikutnya...");
      const ok = await NextRound();
      if (!ok) {
        setStatus("Gagal membuat ronde berikutnya");
        return;
      }
      setStatus("Ronde berikutnya berhasil dibuat");
      await refreshRoundAndStandings();
    } catch (error) {
      console.error("Error creating next round:", error);
      setStatus("Error saat membuat ronde berikutnya");
    }
  };

  const recordResult = async (tableNumber, result) => {
    try {
      setStatus(`Mencatat hasil meja ${tableNumber}...`);
      const ok = await RecordResult(tableNumber, result);
      if (!ok) {
        setStatus("Gagal mencatat hasil");
        return;
      }
      setStatus("Hasil berhasil dicatat");
      await refreshRoundAndStandings();
    } catch (error) {
      console.error("Error recording result:", error);
      setStatus("Error saat mencatat hasil");
    }
  };

  const refreshRoundAndStandings = async () => {
    try {
      const round = await GetCurrentRound();
      setCurrentRound(round);

      const ps = await GetPlayers();
      const map = {};
      (ps || []).forEach((p) => (map[p.id] = p.name));
      setIdToName(map);

      // Sort standings
      ps.sort((a, b) => {
        if (a.score !== b.score) return b.score - a.score;
        if (a.buchholz !== b.buchholz) return b.buchholz - a.buchholz;
        if (a.rating !== b.rating) return b.rating - a.rating;
        return a.name.localeCompare(b.name);
      });
      setStandings(ps);
    } catch (error) {
      console.error("Error refreshing data:", error);
      setStatus("Error saat memperbarui data");
    }
  };

  const getResultButtonStyle = (result, isSelected = false) => {
    const baseStyle = "px-3 py-1 text-sm font-medium transition-colors border";

    if (isSelected) {
      return `${baseStyle} bg-black border-black text-white hover:bg-gray-800`;
    } else {
      return `${baseStyle} bg-white border-black text-black hover:bg-gray-100`;
    }
  };

  return (
    <div className="min-h-screen bg-white text-black flex flex-col relative">
      <div className="absolute inset-0 opacity-5">
        <div className="grid grid-cols-8 h-full">
          {Array.from({ length: 64 }, (_, i) => (
            <div
              key={i}
              className={`${
                Math.floor(i / 8) % 2 === i % 2 ? "bg-black" : "bg-white"
              }`}
            />
          ))}
        </div>
      </div>

      {/* Header */}
      <header className="relative z-10 p-6 border-b border-gray-200">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-3">
            <button
              onClick={() => navigate("/home")}
              className="text-3xl hover:scale-110 transition-transform"
            >
              ♔
            </button>
            <h1 className="text-2xl font-bold tracking-wider">
              XCHESS - Pairing
            </h1>
          </div>
          <nav className="flex space-x-6">
            <button className="text-black text-xl cursor-pointer" title="Login">
              👤
            </button>
            <button
              className="text-black text-xl cursor-pointer"
              title="Settings"
            >
              ⚙️
            </button>
          </nav>
        </div>
      </header>

      {/* Status Bar */}
      {status && (
        <div className="relative z-10 bg-blue-50 border-b border-blue-200 px-6 py-3">
          <p className="text-blue-800 text-sm">{status}</p>
        </div>
      )}

      {/* Tab Navigation */}
      <div className="relative z-10 border-b border-gray-200">
        <div className="flex space-x-8 px-6">
          <button
            onClick={() => setActiveTab("setup")}
            className={`py-4 px-2 border-b-2 font-medium text-sm transition-colors ${
              activeTab === "setup"
                ? "border-black text-black"
                : "border-transparent text-gray-500 hover:text-gray-700"
            }`}
          >
            ♔ Setup Turnamen
          </button>
          <button
            onClick={() => setActiveTab("pairing")}
            className={`py-4 px-2 border-b-2 font-medium text-sm transition-colors ${
              activeTab === "pairing"
                ? "border-black text-black"
                : "border-transparent text-gray-500 hover:text-gray-700"
            }`}
          >
            ♕ Pairing & Hasil
          </button>
          <button
            onClick={() => setActiveTab("standings")}
            className={`py-4 px-2 border-b-2 font-medium text-sm transition-colors ${
              activeTab === "standings"
                ? "border-black text-black"
                : "border-transparent text-gray-500 hover:text-gray-700"
            }`}
          >
            ♖ Klasemen
          </button>
        </div>
      </div>

      {/* Main Content */}
      <main className="flex-1 relative z-10 p-6">
        {/* Setup Tab */}
        {activeTab === "setup" && (
          <div className="max-w-4xl mx-auto">
            <div className="text-center mb-8">
              <h2 className="text-3xl font-bold mb-4">Setup Turnamen Baru</h2>
              <p className="text-gray-600">
                Pilih pemain yang akan berpartisipasi dalam turnamen
              </p>
            </div>

            <div className="bg-white border-2 border-black p-6">
              <h3 className="text-xl font-bold mb-4">
                Pilih Peserta ({selectedIDs.length} dipilih)
              </h3>

              {players.length === 0 ? (
                <p className="text-gray-500 text-center py-8">
                  Tidak ada pemain tersedia
                </p>
              ) : (
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 mb-6">
                  {players.map((player) => (
                    <label
                      key={player.id}
                      className={`flex items-center p-4 border-2 cursor-pointer transition-all ${
                        selectedIDs.includes(player.id)
                          ? "border-black bg-gray-50"
                          : "border-gray-200 hover:border-gray-300"
                      }`}
                    >
                      <input
                        type="checkbox"
                        checked={selectedIDs.includes(player.id)}
                        onChange={() => toggleSelect(player.id)}
                        className="mr-3"
                      />
                      <div>
                        <div className="font-medium">{player.name}</div>
                        <div className="text-sm text-gray-500">
                          Rating: {player.rating}
                        </div>
                      </div>
                    </label>
                  ))}
                </div>
              )}

              <div className="text-center">
                <button
                  onClick={initTournament}
                  disabled={selectedIDs.length < 2}
                  className={`px-8 py-3 font-medium transition-all ${
                    selectedIDs.length < 2
                      ? "bg-gray-300 text-gray-500 cursor-not-allowed"
                      : "bg-black text-white hover:bg-gray-800"
                  }`}
                >
                  Mulai Turnamen ({selectedIDs.length} pemain)
                </button>
              </div>
            </div>
          </div>
        )}

        {/* Pairing Tab */}
        {activeTab === "pairing" && (
          <div className="max-w-6xl mx-auto">
            <div className="text-center mb-8">
              <h2 className="text-3xl font-bold mb-4">
                {currentRound
                  ? `Ronde ${currentRound.round_number}`
                  : "Pairing Turnamen"}
              </h2>
              <p className="text-gray-600">
                Kelola pairing dan catat hasil pertandingan
              </p>
            </div>

            {!currentRound ||
            !currentRound.matches ||
            currentRound.matches.length === 0 ? (
              <div className="text-center py-12">
                <div className="text-6xl mb-4">♔</div>
                <p className="text-gray-500 mb-4">Belum ada pairing tersedia</p>
                <p className="text-sm text-gray-400">
                  Setup turnamen terlebih dahulu untuk membuat pairing
                </p>
              </div>
            ) : (
              <div>
                <div className="bg-white border-2 border-black overflow-hidden">
                  <div className="bg-black text-white p-4">
                    <h3 className="text-lg font-bold">
                      Ronde {currentRound.round_number} - Pairing
                    </h3>
                  </div>

                  <div className="overflow-x-auto">
                    <table className="w-full">
                      <thead className="bg-gray-50">
                        <tr>
                          <th className="px-4 py-3 text-left font-medium">
                            Meja
                          </th>
                          <th className="px-4 py-3 text-left font-medium">
                            Putih
                          </th>
                          <th className="px-4 py-3 text-left font-medium">
                            Hitam
                          </th>
                          <th className="px-4 py-3 text-center font-medium">
                            Hasil
                          </th>
                        </tr>
                      </thead>
                      <tbody>
                        {currentRound.matches.map((match, index) => (
                          <tr
                            key={match.table_number}
                            className={
                              index % 2 === 0 ? "bg-white" : "bg-gray-50"
                            }
                          >
                            <td className="px-4 py-4 font-medium">
                              #{match.table_number}
                            </td>
                            <td className="px-4 py-4">
                              <div className="flex items-center">
                                <span className="text-2xl mr-2">♔</span>
                                {idToName[match.white_id] || match.white_id}
                              </div>
                            </td>
                            <td className="px-4 py-4">
                              <div className="flex items-center">
                                <span className="text-2xl mr-2">♚</span>
                                {match.player_b_id === "BYE"
                                  ? "BYE"
                                  : idToName[match.black_id] || match.black_id}
                              </div>
                            </td>
                            <td className="px-4 py-4">
                              <div className="flex justify-center space-x-2">
                                {match.player_b_id === "BYE" ? (
                                  <button
                                    onClick={() =>
                                      recordResult(match.table_number, "BYE_A")
                                    }
                                    className={getResultButtonStyle(
                                      "BYE_A",
                                      match.result === "BYE_A"
                                    )}
                                  >
                                    Apply Bye
                                  </button>
                                ) : (
                                  <>
                                    <button
                                      onClick={() =>
                                        recordResult(
                                          match.table_number,
                                          "A_WIN"
                                        )
                                      }
                                      className={getResultButtonStyle(
                                        "A_WIN",
                                        match.result === "A_WIN"
                                      )}
                                    >
                                      1-0
                                    </button>
                                    <button
                                      onClick={() =>
                                        recordResult(match.table_number, "DRAW")
                                      }
                                      className={getResultButtonStyle(
                                        "DRAW",
                                        match.result === "DRAW"
                                      )}
                                    >
                                      ½-½
                                    </button>
                                    <button
                                      onClick={() =>
                                        recordResult(
                                          match.table_number,
                                          "B_WIN"
                                        )
                                      }
                                      className={getResultButtonStyle(
                                        "B_WIN",
                                        match.result === "B_WIN"
                                      )}
                                    >
                                      0-1
                                    </button>
                                  </>
                                )}
                              </div>
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                </div>

                <div className="text-center mt-6">
                  <button
                    onClick={nextRound}
                    className="bg-black text-white px-8 py-3 font-medium hover:bg-gray-800 transition-colors"
                  >
                    Generate Ronde Berikutnya
                  </button>
                </div>
              </div>
            )}
          </div>
        )}

        {/* Standings Tab */}
        {activeTab === "standings" && (
          <div className="max-w-4xl mx-auto">
            <div className="text-center mb-8">
              <h2 className="text-3xl font-bold mb-4">Klasemen Turnamen</h2>
              <p className="text-gray-600">
                Peringkat pemain berdasarkan skor dan tiebreak
              </p>
            </div>

            {standings.length === 0 ? (
              <div className="text-center py-12">
                <div className="text-6xl mb-4">♖</div>
                <p className="text-gray-500">Belum ada data klasemen</p>
              </div>
            ) : (
              <div className="bg-white border-2 border-black overflow-hidden">
                <div className="bg-black text-white p-4">
                  <h3 className="text-lg font-bold">Klasemen Sementara</h3>
                </div>

                <div className="overflow-x-auto">
                  <table className="w-full">
                    <thead className="bg-gray-50">
                      <tr>
                        <th className="px-4 py-3 text-left font-medium">
                          Peringkat
                        </th>
                        <th className="px-4 py-3 text-left font-medium">
                          Nama
                        </th>
                        <th className="px-4 py-3 text-center font-medium">
                          Poin
                        </th>
                        <th className="px-4 py-3 text-center font-medium">
                          Dari Round
                        </th>
                        <th className="px-4 py-3 text-center font-medium">
                          Buchholz
                        </th>
                        <th className="px-4 py-3 text-center font-medium">
                          Rating
                        </th>
                        <th className="px-4 py-3 text-center font-medium">
                          Status
                        </th>
                      </tr>
                    </thead>
                    <tbody>
                      {standings.map((player, index) => (
                        <tr
                          key={player.id}
                          className={
                            index % 2 === 0 ? "bg-white" : "bg-gray-50"
                          }
                        >
                          <td className="px-4 py-4 font-bold text-lg">
                            #{index + 1}
                          </td>
                          <td className="px-4 py-4 font-medium">
                            {player.name}
                          </td>
                          <td className="px-4 py-4 text-center font-bold text-lg">
                            {player.score}
                          </td>
                          <td className="px-4 py-4 text-center text-sm text-gray-600">
                            {tournamentInfo
                              ? `${tournamentInfo.current_round} round${
                                  tournamentInfo.current_round > 1 ? "s" : ""
                                }`
                              : "N/A"}
                          </td>
                          <td className="px-4 py-4 text-center">
                            {player.buchholz.toFixed(1)}
                          </td>
                          <td className="px-4 py-4 text-center">
                            {player.rating}
                          </td>
                          <td className="px-4 py-4 text-center">
                            <div className="flex flex-col items-center space-y-1">
                              {player.has_bye && (
                                <span className="px-2 py-1 bg-yellow-100 text-yellow-800 text-xs rounded">
                                  BYE
                                </span>
                              )}
                              {player.color_history && (
                                <span className="text-xs text-gray-500">
                                  {player.color_history
                                    .split("")
                                    .map((color, i) => (
                                      <span
                                        key={i}
                                        className={
                                          color === "W"
                                            ? "text-gray-800"
                                            : "text-gray-400"
                                        }
                                      >
                                        {color === "W" ? "♔" : "♚"}
                                      </span>
                                    ))}
                                </span>
                              )}
                            </div>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </div>
            )}
          </div>
        )}
      </main>

      {/* Footer */}
      <footer className="relative z-10 text-center py-6 border-t border-gray-200">
        <div className="text-sm text-black font-bold">
          maintenance by kewr digital
        </div>
      </footer>
    </div>
  );
}

export default Pairing;
