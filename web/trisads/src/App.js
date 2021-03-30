import './App.css';
const api = require('./pb/trisads/api/v1alpha1/api_grpc_web_pb');

function App() {
  const getStatus = () => {
    const client = new api.TRISADirectoryClient("https://proxy.vaspdirectory.net");
    const req = new api.StatusRequest();
    req.setCommonName("api.alice.vaspbot.net");
    client.status(req, {}, (err, rep) => {
      if (err || !rep) {
        console.log(err);
        return
      }
      console.log(rep);
    });
  };

  return (
    <div className="App">
      <header className="App-header">
        <h1>TRISA Directory Service</h1>
        <button onClick={() => getStatus()}>Status</button>
      </header>
    </div>
  );
}

export default App;
