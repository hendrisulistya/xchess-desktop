import React, { useState } from "react";
import { CheckAdminCredentials } from "../../wailsjs/go/main/App";
import { useNavigate } from "react-router-dom";

function Login({ onSuccess }) {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [message, setMessage] = useState("Please login to continue");
  const navigate = useNavigate();

  const handleLogin = () => {
    setMessage("Checking credentials...");
    CheckAdminCredentials(username, password)
      .then((result) => {
        if (result === true) {
          setMessage("Login successful!");
          onSuccess && onSuccess(username);
          // Navigate to the tournament page
          navigate("/home", { replace: true });
        } else {
          setMessage("Invalid username or password. Please try again.");
        }
      })
      .catch((err) => {
        console.error("Login error:", err);
        setMessage("An error occurred during login. Please try again.");
      });
  };

  const handleKeyPress = (e) => {
    if (e.key === "Enter") {
      handleLogin();
    }
  };

  return (
    <div>
      <div id="result" className="result">
        {message}
      </div>
      <div id="login-form" className="input-box">
        <input
          id="username"
          className="input"
          placeholder="Username"
          onChange={(e) => setUsername(e.target.value)}
          onKeyPress={handleKeyPress}
          autoComplete="off"
          type="text"
        />
        <input
          id="password"
          className="input"
          placeholder="Password"
          onChange={(e) => setPassword(e.target.value)}
          onKeyPress={handleKeyPress}
          type="password"
        />
        <button className="btn" onClick={handleLogin}>
          Login
        </button>
      </div>
    </div>
  );
}

export default Login;
