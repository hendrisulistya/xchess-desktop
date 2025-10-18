import { useNavigate } from "react-router";

function Home() {
  const navigate = useNavigate();

  const handleCreateTournament = () => {
    navigate("/create-tournament");
  };

  return (
    <div className="min-h-screen bg-white text-black flex flex-col relative">
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

      {/* Hero Section */}
      <main className="flex-1 relative z-10 flex flex-col items-center justify-center px-6">
        <div className="text-center max-w-4xl">
          {/* Chess Pieces Icons */}
          <div className="flex justify-center space-x-4 text-6xl mb-8">
            <span>♔</span>
            <span>♕</span>
            <span>♖</span>
            <span>♗</span>
            <span>♘</span>
            <span>♙</span>
          </div>

          {/* Main Title */}
          <h1 className="text-5xl font-bold mb-4 tracking-wider">
            Tournament Management
          </h1>

          {/* Subtitle */}
          <p className="text-xl text-gray-600 mb-12 font-light max-w-2xl mx-auto">
            Kelola turnamen catur Anda dengan sistem pairing otomatis, tracking
            skor real-time, dan manajemen pemain yang mudah
          </p>

          {/* Action Buttons */}
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6 max-w-3xl mx-auto">
            {/* Create Tournament Card */}
            <div
              className="bg-white border-2 border-black p-6 hover:shadow-lg transition-all duration-300 group cursor-pointer"
              onClick={handleCreateTournament}
            >
              <div className="text-4xl mb-4 group-hover:scale-110 transition-transform">
                ♔
              </div>
              <h3 className="text-xl font-bold mb-2">Buat Turnamen</h3>
              <p className="text-gray-600 text-sm mb-4">
                Mulai turnamen baru dengan sistem Swiss pairing
              </p>
              <div className="bg-black text-white px-4 py-2 text-sm font-medium group-hover:bg-gray-800 transition-colors">
                Mulai Sekarang
              </div>
            </div>

            {/* View Tournaments Card - Coming Soon */}
            <div className="bg-gray-100 border-2 border-gray-300 p-6 transition-all duration-300 group cursor-not-allowed opacity-60">
              <div className="text-4xl mb-4 text-gray-400">♕</div>
              <h3 className="text-xl font-bold mb-2 text-gray-500">
                Lihat Turnamen
              </h3>
              <p className="text-gray-400 text-sm mb-4">
                Pantau progress dan hasil turnamen yang sedang berjalan
              </p>
              <div className="bg-gray-400 text-white px-4 py-2 text-sm font-medium">
                Coming Soon
              </div>
            </div>

            {/* Manage Players Card - Coming Soon */}
            <div className="bg-gray-100 border-2 border-gray-300 p-6 transition-all duration-300 group cursor-not-allowed opacity-60">
              <div className="text-4xl mb-4 text-gray-400">♖</div>
              <h3 className="text-xl font-bold mb-2 text-gray-500">
                Kelola Pemain
              </h3>
              <p className="text-gray-400 text-sm mb-4">
                Tambah, edit, dan kelola database pemain catur
              </p>
              <div className="bg-gray-400 text-white px-4 py-2 text-sm font-medium">
                Coming Soon
              </div>
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

export default Home;
