import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router";
import {
  NextRound,
  GetCurrentRound,
  RecordResult,
  GetPlayers,
  GetTournamentInfo,
} from "../../wailsjs/go/main/App";
import Navbar from "../components/Navbar";

function Pairing() {
  const navigate = useNavigate();
  const [currentRound, setCurrentRound] = useState(null);
  const [standings, setStandings] = useState([]);
  const [tournamentInfo, setTournamentInfo] = useState(null);
  const [status, setStatus] = useState("");
  const [idToName, setIdToName] = useState({});
  const [activeTab, setActiveTab] = useState("pairing");

  useEffect(() => {
    loadInitialData();
  }, []);

  const handleTabChange = (tab) => {
    setActiveTab(tab);
    if (tab === "setup") {
      navigate("/setup-tournament");
    } else if (tab === "pairing") {
      navigate("/pairing");
    } else if (tab === "standings") {
      navigate("/klasemen");
    }
  };

  const loadInitialData = async () => {
    try {
      const tournamentData = await GetTournamentInfo();
      if (tournamentData) {
        setTournamentInfo(tournamentData);
        await refreshRoundAndStandings();
      } else {
        setStatus(
          "Tidak ada turnamen aktif. Silakan setup turnamen terlebih dahulu."
        );
      }
    } catch (error) {
      console.error("Error loading initial data:", error);
      setStatus("Error loading data");
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

      {/* Header with Logo */}
      <Navbar
        activeTab={activeTab}
        setActiveTab={handleTabChange}
        showTabs={true}
      />

      {/* Status Bar */}
      {status && (
        <div className="relative z-10 bg-blue-50 border-b border-blue-200 px-6 py-3">
          <p className="text-blue-800 text-sm">{status}</p>
        </div>
      )}

      {/* Main Content */}
      <main className="flex-1 relative z-10 p-6">
        <div className="max-w-6xl mx-auto">
          <div className="text-center mb-8">
            <h2 className="text-3xl font-bold mb-2">
              {tournamentInfo?.title || "Pairing Turnamen"}
            </h2>
            <p className="text-gray-600">
              {tournamentInfo?.description ||
                "Kelola pairing dan catat hasil pertandingan"}
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
              <button
                onClick={() => navigate("/setup-tournament")}
                className="mt-4 bg-black text-white px-6 py-2 hover:bg-gray-800 transition-colors"
              >
                Setup Turnamen
              </button>
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
                        <th className="px-4 py-3 text-center font-medium">
                          Poin
                        </th>
                        <th className="px-4 py-3 text-left font-medium">
                          Hitam
                        </th>
                        <th className="px-4 py-3 text-center font-medium">
                          Poin
                        </th>
                        <th className="px-4 py-3 text-center font-medium">
                          Hasil
                        </th>
                      </tr>
                    </thead>
                    <tbody>
                      {currentRound.matches.map((match, index) => {
                        const whitePlayer = standings.find(
                          (p) => p.id === match.white_id
                        );
                        const blackPlayer = standings.find(
                          (p) => p.id === match.black_id
                        );

                        return (
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
                            <td className="px-4 py-4 text-center font-bold">
                              {whitePlayer ? whitePlayer.score : 0}
                            </td>
                            <td className="px-4 py-4">
                              <div className="flex items-center">
                                <span className="text-2xl mr-2">♚</span>
                                {match.player_b_id === "BYE"
                                  ? "BYE"
                                  : idToName[match.black_id] || match.black_id}
                              </div>
                            </td>
                            <td className="px-4 py-4 text-center font-bold">
                              {match.player_b_id === "BYE"
                                ? "-"
                                : blackPlayer
                                ? blackPlayer.score
                                : 0}
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
                        );
                      })}
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
