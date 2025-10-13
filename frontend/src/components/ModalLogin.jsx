import React, { useState } from "react";
import { CheckAdminCredentials } from "../../wailsjs/go/main/App";
import { useNavigate } from "react-router";

function ModalLogin({ isOpen, onClose, onSuccess }) {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [message, setMessage] = useState(
    "Masukkan kredensial admin untuk melanjutkan"
  );
  const [isLoading, setIsLoading] = useState(false);
  const navigate = useNavigate();

  // Check if Wails runtime is available
  const isWailsAvailable = () => {
    return (
      typeof window !== "undefined" &&
      window.go &&
      window.go.main &&
      window.go.main.App &&
      window.go.main.App.CheckAdminCredentials
    );
  };

  const handleLogin = () => {
    if (!username || !password) {
      setMessage("Username dan password harus diisi");
      return;
    }

    // Check if Wails runtime is available
    if (!isWailsAvailable()) {
      setMessage("Aplikasi belum siap. Silakan tunggu sebentar dan coba lagi.");
      console.error("Wails runtime not available");
      return;
    }

    setIsLoading(true);
    setMessage("Memeriksa kredensial...");

    CheckAdminCredentials(username, password)
      .then((result) => {
        if (result === true) {
          setMessage("Login berhasil!");
          onSuccess && onSuccess(username);
          // Navigate to the create tournament page
          navigate("/create-tournament", { replace: true });
          onClose();
        } else {
          setMessage("Username atau password salah. Silakan coba lagi.");
        }
      })
      .catch((err) => {
        console.error("Login error:", err);
        setMessage("Terjadi kesalahan saat login. Silakan coba lagi.");
      })
      .finally(() => {
        setIsLoading(false);
      });
  };

  const handleKeyPress = (e) => {
    if (e.key === "Enter") {
      handleLogin();
    }
  };

  const handleClose = () => {
    setUsername("");
    setPassword("");
    setMessage("Masukkan kredensial admin untuk melanjutkan");
    setIsLoading(false);
    onClose();
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      {/* Modal Backdrop */}
      <div className="absolute inset-0" onClick={handleClose}></div>

      {/* Modal Content */}
      <div className="relative bg-white border-4 border-black max-w-md w-full mx-4 z-10">
        {/* Chess Board Pattern Background */}
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

        {/* Modal Header */}
        <div className="relative z-10 p-6 border-b-2 border-black">
          <div className="flex justify-between items-center">
            <div className="flex items-center space-x-3">
              <div className="text-2xl">♔</div>
              <h2 className="text-xl font-bold tracking-wider">LOGIN ADMIN</h2>
            </div>
            <button
              onClick={handleClose}
              className="text-2xl font-bold hover:bg-gray-100 w-8 h-8 flex items-center justify-center transition-all duration-300"
            >
              ×
            </button>
          </div>
        </div>

        {/* Modal Body */}
        <div className="relative z-10 p-6">
          {/* Runtime Status Check */}
          {!isWailsAvailable() && (
            <div className="mb-4 p-3 bg-yellow-100 border-2 border-yellow-300 text-yellow-800 text-sm">
              <strong>Peringatan:</strong> Aplikasi belum sepenuhnya dimuat.
              Pastikan Anda menjalankan aplikasi melalui Wails.
            </div>
          )}

          {/* Message */}
          <div className="mb-6 text-center">
            <p
              className={`text-sm font-medium ${
                message.includes("berhasil")
                  ? "text-green-700"
                  : message.includes("salah") || message.includes("kesalahan")
                  ? "text-red-700"
                  : "text-gray-700"
              }`}
            >
              {message}
            </p>
          </div>

          {/* Login Form */}
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-black mb-2">
                Username
              </label>
              <input
                type="text"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                onKeyPress={handleKeyPress}
                placeholder="Masukkan username"
                className="w-full px-4 py-3 border-2 border-black focus:outline-none focus:ring-0 focus:border-gray-600 transition-all duration-300"
                autoComplete="off"
                disabled={isLoading || !isWailsAvailable()}
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-black mb-2">
                Password
              </label>
              <input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                onKeyPress={handleKeyPress}
                placeholder="Masukkan password"
                className="w-full px-4 py-3 border-2 border-black focus:outline-none focus:ring-0 focus:border-gray-600 transition-all duration-300"
                disabled={isLoading || !isWailsAvailable()}
              />
            </div>
          </div>

          {/* Action Buttons */}
          <div className="flex space-x-3 mt-8">
            <button
              onClick={handleClose}
              className="flex-1 bg-white text-black px-4 py-3 font-medium border-2 border-black hover:bg-gray-100 transition-all duration-300"
              disabled={isLoading}
            >
              Batal
            </button>
            <button
              onClick={handleLogin}
              disabled={isLoading || !isWailsAvailable()}
              className="flex-1 bg-black text-white px-4 py-3 font-medium border-2 border-black hover:bg-gray-800 transition-all duration-300 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {isLoading ? "Memproses..." : "Login"}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}

export default ModalLogin;
