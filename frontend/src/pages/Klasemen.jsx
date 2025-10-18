import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router";
import {
  GetPlayers,
  GetTournamentInfo,
  SaveStandingsToPDF,
} from "../../wailsjs/go/main/App";
import Navbar from "../components/Navbar";

function Klasemen() {
  const navigate = useNavigate();
  const [standings, setStandings] = useState([]);
  const [tournamentInfo, setTournamentInfo] = useState(null);
  const [status, setStatus] = useState("");
  const [activeTab, setActiveTab] = useState("standings");

  useEffect(() => {
    loadStandings();
  }, []);

  const loadStandings = async () => {
    try {
      const tournamentData = await GetTournamentInfo();
      setTournamentInfo(tournamentData);

      const ps = await GetPlayers();

      ps.sort((a, b) => {
        if (a.score !== b.score) return b.score - a.score;
        if (a.buchholz !== b.buchholz) return b.buchholz - a.buchholz;
        // Remove rating sort since we don't have rating anymore
        return a.name.localeCompare(b.name);
      });
      setStandings(ps);
    } catch (error) {
      console.error("Error loading standings:", error);
      setStatus("Error saat memuat klasemen");
    }
  };

  const handleTabChange = (tabId) => {
    setActiveTab(tabId);
    switch (tabId) {
      case "setup":
        navigate("/setup-tournament");
        break;
      case "pairing":
        navigate("/pairing");
        break;
      case "standings":
        break;
      default:
        break;
    }
  };

  const exportStandingsToPDF = async () => {
    try {
      if (standings.length === 0) {
        setStatus("Tidak ada data klasemen untuk diekspor");
        return;
      }

      setStatus("Menyimpan klasemen ke PDF...");
      console.log("Starting PDF save for standings");

      try {
        // Use the save function that saves directly to Desktop
        const filePath = await SaveStandingsToPDF();

        console.log("PDF saved to:", filePath);
        setStatus(
          `PDF berhasil disimpan ke Desktop: ${filePath.split("/").pop()}`
        );
      } catch (backendError) {
        console.error("Backend error:", backendError);
        setStatus(`Error saat menyimpan PDF: ${backendError.toString()}`);
      }
    } catch (error) {
      console.error("Error saving standings to PDF:", error);
      setStatus("Error saat menyimpan klasemen ke PDF");
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
        <div className="max-w-4xl mx-auto">
          <div className="flex justify-between items-center mb-8">
            <div className="text-center flex-1">
              <h2 className="text-3xl font-bold mb-4">Klasemen Turnamen</h2>
              <p className="text-gray-600">
                Peringkat pemain berdasarkan skor dan tiebreak
              </p>
            </div>
            <button
              onClick={exportStandingsToPDF}
              className="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded"
            >
              Export PDF
            </button>
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
                      <th className="px-4 py-3 text-left font-medium">Nama</th>
                      <th className="px-4 py-3 text-center font-medium">
                        Poin
                      </th>
                      <th className="px-4 py-3 text-center font-medium">
                        Buchholz
                      </th>
                      <th className="px-4 py-3 text-center font-medium">
                        Progressive Score
                      </th>
                      <th className="px-4 py-3 text-center font-medium">
                        Club / Domisili
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    {standings.map((player, index) => (
                      <tr
                        key={player.id}
                        className={index % 2 === 0 ? "bg-white" : "bg-gray-50"}
                      >
                        <td className="px-4 py-4 font-bold text-lg">
                          #{index + 1}
                        </td>
                        <td className="px-4 py-4 font-medium">{player.name}</td>
                        <td className="px-4 py-4 text-center font-bold text-lg">
                          {player.score}
                        </td>
                        <td className="px-4 py-4 text-center">
                          {player.buchholz.toFixed(1)}
                        </td>
                        <td className="px-4 py-4 text-center">
                          {player.progressive_score
                            ? player.progressive_score.toFixed(1)
                            : "0.0"}
                        </td>
                        <td className="px-4 py-4 text-center">{player.club}</td>
                        <td className="px-4 py-4 text-center">
                          <div className="flex flex-col items-center space-y-1">
                            {player.has_bye && (
                              <span className="px-2 py-1 bg-yellow-100 text-yellow-800 text-xs rounded">
                                BYE
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

export default Klasemen;
