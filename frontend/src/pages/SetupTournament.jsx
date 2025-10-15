import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router";
import {
  ListPlayers,
  InitTournamentWithPlayerIDs,
  NextRound,
  GetTournamentInfo,
  AddPlayer,
} from "../../wailsjs/go/main/App";
import Navbar from "../components/Navbar";

function SetupTournament() {
  const navigate = useNavigate();
  const [players, setPlayers] = useState([]);
  const [selectedIDs, setSelectedIDs] = useState([]);
  const [status, setStatus] = useState("");
  const [activeTab, setActiveTab] = useState("setup");

  // State untuk form tambah pemain
  const [showAddPlayerForm, setShowAddPlayerForm] = useState(false);
  // State untuk form tambah pemain
  const [newPlayer, setNewPlayer] = useState({
    name: "",
    club: "",
  });
  const [isAddingPlayer, setIsAddingPlayer] = useState(false);

  useEffect(() => {
    loadPlayers();
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

  const loadPlayers = async () => {
    try {
      const playerList = await ListPlayers();
      setPlayers(playerList || []);
    } catch (error) {
      console.error("Error loading players:", error);
      setStatus("Error loading players");
    }
  };

  // Fungsi untuk menambah pemain baru
  const handleAddPlayer = async (e) => {
    e.preventDefault();

    if (!newPlayer.name.trim()) {
      setStatus("Nama pemain tidak boleh kosong");
      return;
    }

    try {
      setIsAddingPlayer(true);
      setStatus("Menambahkan pemain baru...");

      const playerID = await AddPlayer(
        newPlayer.name.trim(),
        newPlayer.club.trim()
      );

      if (playerID) {
        setStatus(`Pemain ${newPlayer.name} berhasil ditambahkan`);

        // Reset form
        setNewPlayer({ name: "", club: "" });
        setShowAddPlayerForm(false);

        // Reload players list
        await loadPlayers();

        // Clear status after 3 seconds
        setTimeout(() => setStatus(""), 3000);
      } else {
        setStatus("Gagal menambahkan pemain");
      }
    } catch (error) {
      console.error("Error adding player:", error);
      setStatus("Error saat menambahkan pemain: " + error.message);
    } finally {
      setIsAddingPlayer(false);
    }
  };

  const handleCancelAddPlayer = () => {
    setNewPlayer({ name: "", rating: 1200 });
    setShowAddPlayerForm(false);
    setStatus("");
  };

  const toggleSelect = (id) => {
    setSelectedIDs((prev) =>
      prev.includes(id) ? prev.filter((x) => x !== id) : [...prev, id]
    );
  };

  const selectAll = () => {
    if (selectedIDs.length === players.length) {
      setSelectedIDs([]);
    } else {
      setSelectedIDs(players.map((player) => player.id));
    }
  };

  const initTournament = async () => {
    if (selectedIDs.length < 2) {
      setStatus("Pilih minimal 2 pemain untuk memulai turnamen");
      return;
    }

    try {
      setStatus("Menginisialisasi turnamen...");

      let title = "Turnamen Baru";
      let description = "Turnamen catur dengan sistem Swiss";

      try {
        const existingTournament = await GetTournamentInfo();
        if (existingTournament && existingTournament.title) {
          title = existingTournament.title;
          description = existingTournament.description || description;
        }
      } catch (error) {
        console.log("No existing tournament found, using defaults");
      }

      const ok = await InitTournamentWithPlayerIDs(
        title,
        description,
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

      setStatus("Ronde 1 berhasil dibuat. Mengarahkan ke halaman pairing...");

      setTimeout(() => {
        navigate("/pairing");
      }, 1500);
    } catch (error) {
      console.error("Error initializing tournament:", error);
      setStatus("Error saat menginisialisasi turnamen");
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
          <div className="text-center mb-8">
            <h2 className="text-3xl font-bold mb-4">Setup Turnamen Baru</h2>
            <p className="text-gray-600">
              Pilih pemain yang akan berpartisipasi dalam turnamen
            </p>
          </div>

          {/* Form Tambah Pemain Baru */}
          <div className="bg-white border-2 border-black p-6 mb-6">
            <div className="flex justify-between items-center mb-4">
              <h3 className="text-xl font-bold">Tambah Pemain Baru</h3>
              <button
                onClick={() => setShowAddPlayerForm(!showAddPlayerForm)}
                className="px-4 py-2 text-sm font-medium border-2 border-black bg-white text-black hover:bg-gray-100 transition-colors"
              >
                {showAddPlayerForm ? "Tutup Form" : "Tambah Pemain"}
              </button>
            </div>

            {showAddPlayerForm && (
              <form onSubmit={handleAddPlayer} className="space-y-4">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm font-medium mb-2">
                      Nama Pemain *
                    </label>
                    <input
                      type="text"
                      value={newPlayer.name}
                      onChange={(e) =>
                        setNewPlayer({ ...newPlayer, name: e.target.value })
                      }
                      placeholder="Masukkan nama pemain"
                      className="w-full px-3 py-2 border-2 border-gray-300 focus:border-black focus:outline-none"
                      required
                      disabled={isAddingPlayer}
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium mb-2">
                      Club
                    </label>
                    <input
                      type="text"
                      value={newPlayer.club}
                      onChange={(e) =>
                        setNewPlayer({ ...newPlayer, club: e.target.value })
                      }
                      placeholder="Masukkan nama club"
                      className="w-full px-3 py-2 border-2 border-gray-300 focus:border-black focus:outline-none"
                      disabled={isAddingPlayer}
                    />
                  </div>
                </div>

                <div className="flex space-x-4">
                  <button
                    type="submit"
                    disabled={isAddingPlayer || !newPlayer.name.trim()}
                    className={`px-6 py-2 font-medium transition-all ${
                      isAddingPlayer || !newPlayer.name.trim()
                        ? "bg-gray-300 text-gray-500 cursor-not-allowed"
                        : "bg-black text-white hover:bg-gray-800"
                    }`}
                  >
                    {isAddingPlayer ? "Menambahkan..." : "Tambah Pemain"}
                  </button>
                  <button
                    type="button"
                    onClick={handleCancelAddPlayer}
                    disabled={isAddingPlayer}
                    className="px-6 py-2 font-medium border-2 border-gray-300 bg-white text-gray-700 hover:bg-gray-50 transition-colors"
                  >
                    Batal
                  </button>
                </div>
              </form>
            )}
          </div>

          {/* Pilih Peserta */}
          <div className="bg-white border-2 border-black p-6">
            <div className="flex justify-between items-center mb-4">
              <h3 className="text-xl font-bold">
                Pilih Peserta ({selectedIDs.length} dipilih)
              </h3>
              <button
                onClick={selectAll}
                className="px-4 py-2 text-sm font-medium border-2 border-black bg-white text-black hover:bg-gray-100 transition-colors"
              >
                {selectedIDs.length === players.length
                  ? "Deselect All"
                  : "Select All"}
              </button>
            </div>

            {players.length === 0 ? (
              <div className="text-center py-8">
                <p className="text-gray-500 mb-4">Tidak ada pemain tersedia</p>
                <p className="text-sm text-gray-400">
                  Tambahkan pemain baru menggunakan form di atas
                </p>
              </div>
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
                        Club: {player.club || "Tidak ada club"}
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

export default SetupTournament;
