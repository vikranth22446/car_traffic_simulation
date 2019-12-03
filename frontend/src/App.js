import React from 'react';
import "./App.css";
import SimulationHandler from "./SimulationHandler";


function App() {

    return (
        <div className="App">
            <div className="Card title">
                <div className={"titleText"}>EE126 Simulation</div>
                <img className="titleLogo" src={"/static/ee126_logo.png"}/>
            </div>
            <SimulationHandler/>
        </div>
    );

}

export default App;
