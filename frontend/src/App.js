import React, {Component} from "react";
import "./App.css";


class App extends Component {

    render() {
        const input = '# This is a header\n\nAnd this is a paragraph'
        return (
            <div className="App">
                <div className="Card">
                    <Upload/>
                </div>
                <div className="Section">
                    <Gallery/>
                </div>
            </div>
        );
    }
}

export default App;
