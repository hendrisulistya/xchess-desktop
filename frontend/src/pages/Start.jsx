import React from "react";
import { useNavigate } from "react-router";
import Logo from "../assets/images/xchess.png";

function Start() {
  const navigate = useNavigate();

  const handleLogin = () => {
    navigate("/home");
  };

  return (
    <div className="min-h-screen bg-white text-black flex flex-col items-center justify-center relative">
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

      {/* Main Content */}
      <div className="text-center relative z-10 flex-1 flex flex-col justify-center">
        {/* Logo Image */}
        <div className="mb-6">
          <img src={Logo} alt="XCHESS Logo" className="h-32 w-auto mx-auto" />
        </div>

        {/* Login Button */}
        <button
          onClick={handleLogin}
          className="bg-black text-white px-12 py-4 text-lg font-medium hover:bg-gray-800 transition-all duration-300 border-2 border-black hover:shadow-lg cursor-pointer"
        >
          Mulai
        </button>
      </div>

      {/* Maintenance Credit - Bottom */}
      <div className="text-sm text-black font-bold pb-6 relative z-10">
        maintenance by kewr digital
      </div>
    </div>
  );
}

export default Start;
