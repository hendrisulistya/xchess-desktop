import React from "react";
import logo from "./assets/images/logo-universal.png";
import "./App.css";
import { Routes, Route, Navigate } from "react-router";
import Pairing from "./pages/Pairing";
import Home from "./pages/Home";
import Start from "./pages/Start";
import CreateTournament from "./pages/CreateTournament";

function App() {
  return (
    <div id="App">
      <Routes>
        <Route path="/" element={<Start />} />
        <Route path="/home" element={<Home />} />
        <Route path="/create-tournament" element={<CreateTournament />} />
        <Route path="/pairing" element={<Pairing />} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </div>
  );
}

export default App;
