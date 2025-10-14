import React, { useState } from "react";
import { InitTournament } from "../../wailsjs/go/main/App";
import { useNavigate } from "react-router";
import Logo from "../assets/images/xchess.png";

function CreateTournament() {
  const navigate = useNavigate();
  const [formData, setFormData] = useState({
    title: "",
    description: "",
    rounds_total: 5,
    pairing_system: "Swiss",
    bye_score: 1.0,
  });
  const [isLoading, setIsLoading] = useState(false);
  const [message, setMessage] = useState("");

  const handleInputChange = (e) => {
    const { name, value } = e.target;
    setFormData((prev) => ({
      ...prev,
      [name]:
        name === "rounds_total"
          ? parseInt(value) || 1
          : name === "bye_score"
          ? parseFloat(value) || 0.5
          : value,
    }));
  };

  const handleSubmit = async (e) => {
    e.preventDefault();

    if (!formData.title.trim()) {
      setMessage("Nama tournament harus diisi");
      return;
    }

    if (formData.rounds_total < 1 || formData.rounds_total > 20) {
      setMessage("Jumlah ronde harus antara 1-20");
      return;
    }

    if (formData.bye_score < 0 || formData.bye_score > 1) {
      setMessage("Skor bye harus antara 0-1");
      return;
    }

    setIsLoading(true);
    setMessage("Membuat tournament...");

    try {
      const result = await InitTournament(
        formData.title,
        formData.description.trim() || "Tournament catur dengan sistem Swiss",
        []
      );

      if (result) {
        setMessage("Tournament berhasil dibuat!");
        setTimeout(() => {
          navigate("/pairing");
        }, 2000);
      } else {
        setMessage("Gagal membuat tournament. Silakan coba lagi.");
      }
    } catch (error) {
      console.error("Error creating tournament:", error);
      setMessage("Terjadi kesalahan saat membuat tournament.");
    } finally {
      setIsLoading(false);
    }
  };

  const handleBack = () => {
    navigate("/home");
  };

  return (
    <div className="min-h-screen bg-white text-black relative">
      {/* Chess Board Pattern Background */}
      <div className="absolute inset-0 opacity-3">
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

      {/* Main Content */}
      <div className="relative z-10 p-8">
        {/* Top Navigation Bar */}
        <div className="flex justify-between items-center mb-8">
          <div className="flex items-center space-x-3">
            <img src={Logo} alt="XCHESS Logo" className="h-8 w-auto" />
          </div>

          <div className="flex items-center">
            <button
              onClick={handleBack}
              className="bg-white text-black px-6 py-2 text-sm font-medium border-2 border-black hover:bg-black hover:text-white transition-all duration-300"
            >
              Kembali
            </button>
          </div>
        </div>

        {/* Header */}
        <div className="text-center mb-8">
          <h1 className="text-3xl font-bold mb-2 tracking-wider">
            BUAT TOURNAMENT
          </h1>
          <p className="text-base text-gray-600 font-light">
            Buat tournament catur baru
          </p>
        </div>

        {/* Create Tournament Form */}
        <div className="max-w-2xl mx-auto">
          <div className="bg-white border-4 border-black p-8">
            {/* Message */}
            {message && (
              <div className="mb-6 text-center">
                <p
                  className={`text-sm font-medium ${
                    message.includes("berhasil")
                      ? "text-green-700"
                      : message.includes("Gagal") ||
                        message.includes("kesalahan")
                      ? "text-red-700"
                      : "text-blue-700"
                  }`}
                >
                  {message}
                </p>
              </div>
            )}

            <form onSubmit={handleSubmit} className="space-y-6">
              {/* Tournament Title */}
              <div>
                <label className="block text-lg font-bold text-black mb-3">
                  Nama Tournament
                </label>
                <input
                  type="text"
                  name="title"
                  value={formData.title}
                  onChange={handleInputChange}
                  placeholder="Masukkan nama tournament"
                  className="w-full px-4 py-4 border-2 border-black focus:outline-none focus:ring-0 focus:border-gray-600 transition-all duration-300 text-lg"
                  disabled={isLoading}
                  required
                />
                <p className="text-sm text-gray-600 mt-2">
                  Contoh: Jakarta Chess Championship 2024
                </p>
              </div>

              {/* Tournament Description */}
              <div>
                <label className="block text-lg font-bold text-black mb-3">
                  Deskripsi Tournament
                </label>
                <textarea
                  name="description"
                  value={formData.description}
                  onChange={handleInputChange}
                  placeholder="Masukkan deskripsi tournament (opsional)"
                  rows="3"
                  className="w-full px-4 py-4 border-2 border-black focus:outline-none focus:ring-0 focus:border-gray-600 transition-all duration-300 text-lg resize-none"
                  disabled={isLoading}
                />
                <p className="text-sm text-gray-600 mt-2">
                  Deskripsi singkat tentang tournament ini
                </p>
              </div>

              {/* Number of Rounds */}
              <div>
                <label className="block text-lg font-bold text-black mb-3">
                  Jumlah Ronde
                </label>
                <input
                  type="number"
                  name="rounds_total"
                  value={formData.rounds_total}
                  onChange={handleInputChange}
                  min="1"
                  max="20"
                  className="w-full px-4 py-4 border-2 border-black focus:outline-none focus:ring-0 focus:border-gray-600 transition-all duration-300 text-lg"
                  disabled={isLoading}
                  required
                />
                <p className="text-sm text-gray-600 mt-2">
                  Jumlah ronde tournament (1-20 ronde)
                </p>
              </div>

              {/* Pairing System */}
              <div>
                <label className="block text-lg font-bold text-black mb-3">
                  Sistem Pairing
                </label>
                <select
                  name="pairing_system"
                  value={formData.pairing_system}
                  onChange={handleInputChange}
                  className="w-full px-4 py-4 border-2 border-black focus:outline-none focus:ring-0 focus:border-gray-600 transition-all duration-300 text-lg"
                  disabled={isLoading}
                >
                  <option value="Swiss">Swiss System</option>
                  <option value="Round Robin">Round Robin</option>
                </select>
                <p className="text-sm text-gray-600 mt-2">
                  Sistem pairing yang akan digunakan dalam tournament
                </p>
              </div>

              {/* Bye Score */}
              <div>
                <label className="block text-lg font-bold text-black mb-3">
                  Skor Bye
                </label>
                <input
                  type="number"
                  name="bye_score"
                  value={formData.bye_score}
                  onChange={handleInputChange}
                  min="0"
                  max="1"
                  step="0.5"
                  className="w-full px-4 py-4 border-2 border-black focus:outline-none focus:ring-0 focus:border-gray-600 transition-all duration-300 text-lg"
                  disabled={isLoading}
                  required
                />
                <p className="text-sm text-gray-600 mt-2">
                  Skor yang diberikan untuk bye (biasanya 0.5 atau 1.0)
                </p>
              </div>

              {/* Tournament Info */}
              <div className="bg-gray-50 border-2 border-gray-300 p-4">
                <h3 className="font-bold text-black mb-2">
                  Informasi Tournament:
                </h3>
                <ul className="text-sm text-gray-700 space-y-1">
                  <li>
                    • Tournament akan menggunakan sistem{" "}
                    {formData.pairing_system}
                  </li>
                  <li>• Pemain dapat ditambahkan setelah tournament dibuat</li>
                  <li>• Pairing akan dibuat otomatis setiap ronde</li>
                  <li>• Hasil pertandingan dapat diinput secara real-time</li>
                  <li>• Skor bye: {formData.bye_score} poin</li>
                </ul>
              </div>

              {/* Action Buttons */}
              <div className="flex space-x-4 pt-4">
                <button
                  type="button"
                  onClick={handleBack}
                  className="flex-1 bg-white text-black px-6 py-4 text-lg font-medium border-2 border-black hover:bg-gray-100 transition-all duration-300"
                  disabled={isLoading}
                >
                  Batal
                </button>
                <button
                  type="submit"
                  disabled={isLoading}
                  className="flex-1 bg-black text-white px-6 py-4 text-lg font-medium border-2 border-black hover:bg-gray-800 transition-all duration-300 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {isLoading ? "Membuat..." : "Buat Tournament"}
                </button>
              </div>
            </form>
          </div>
        </div>
      </div>

      {/* Footer Credit */}
      <div className="text-center text-sm text-black font-bold pb-6 relative z-10">
        maintenance by kewr digital
      </div>
    </div>
  );
}

export default CreateTournament;
