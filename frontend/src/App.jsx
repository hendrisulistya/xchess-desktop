import React from "react";
import logo from "./assets/images/logo-universal.png";
import "./App.css";
import Login from "./components/Login";
import { BrowserRouter, Routes, Route, Link, Navigate } from "react-router-dom";
import Tournament from "./components/Tournament";

function App() {
  return (
    <div id="App">
      <img src={logo} id="logo" alt="logo" />
      {/* Use the router provided in main.jsx; do not nest another BrowserRouter here */}
      <Routes>
        <Route path="/" element={<Login />} />
        <Route path="/home" element={<Tournament />} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </div>
  );
}

export default App;
