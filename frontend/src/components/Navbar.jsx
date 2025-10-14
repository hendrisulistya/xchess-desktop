import React from "react";
import Logo from "../assets/images/xchess.png";

function Navbar({ activeTab, setActiveTab, showTabs = true }) {
  const tabs = [
    {
      id: "setup",
      label: "♔ Setup Turnamen",
      icon: "♔",
    },
    {
      id: "pairing",
      label: "♕ Pairing & Hasil",
      icon: "♕",
    },
    {
      id: "standings",
      label: "♖ Klasemen",
      icon: "♖",
    },
  ];

  return (
    <header className="relative z-10 border-b border-gray-200">
      <div className="flex items-center justify-between px-6 py-4">
        <div className="flex items-center space-x-3">
          <img src={Logo} alt="XCHESS Logo" className="h-8 w-auto" />
        </div>

        {/* Navigation Tabs */}
        {showTabs && (
          <div className="flex space-x-8">
            {tabs.map((tab) => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                className={`py-2 px-3 border-b-2 font-medium text-sm transition-colors ${
                  activeTab === tab.id
                    ? "border-black text-black"
                    : "border-transparent text-gray-500 hover:text-gray-700"
                }`}
              >
                {tab.label}
              </button>
            ))}
          </div>
        )}
      </div>
    </header>
  );
}

export default Navbar;
