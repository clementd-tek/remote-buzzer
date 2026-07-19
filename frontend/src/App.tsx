import { Route, Routes } from "react-router-dom";
import { HomePage } from "./pages/HomePage";
import { LobbyPage } from "./pages/LobbyPage";

export default function App() {
  return (
    <Routes>
      <Route path="/" element={<HomePage />} />
      <Route path="/lobby/:id" element={<LobbyPage />} />
      <Route path="*" element={<NotFound />} />
    </Routes>
  );
}

function NotFound() {
  return (
    <div className="container" style={{ padding: "72px 24px", textAlign: "center" }}>
      <h1>404</h1>
      <p>Cette page n'existe pas.</p>
    </div>
  );
}
