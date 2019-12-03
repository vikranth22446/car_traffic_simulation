import React, {Component} from 'react';

class SimulationCar extends Component {
    render() {
        var items = [];
        var value;
        for (var key in this.props.car) {
            value = this.props.car[key];
            items.push({"name": key})
        }
        if (items.length !== 0) {
            return (
                <td>
                    <div>

                        {items.map(item => <div>🚗{item.name}</div>)}
                    </div>
                </td>
            );
        }

        if (this.props.locationState === 0) {
            return <td className={"empty"}>
                <div></div>
            </td>
        }

        return <td>
            <div>&nbsp;&nbsp;&nbsp;&nbsp;</div>
        </td>
    }
}

class Simulation extends Component {
    render() {
        if (this.props.simulating) {
            console.log(this.props.data);
            return <table className={"tableBorder"}>
                {this.props.data.map(row => {
                    return <tr>
                        {row.map(item => {
                            return <SimulationCar locationState={item.state} car={item.cars}/>
                        })}
                        {/*<td key={row.name}>{row.name}</td>*/}
                        {/*<td key={row.id}>{row.id}</td>*/}
                        {/*<td key={row.price}>{row.price}</td>*/}
                    </tr>
                })}
            </table>
        }

        return <div></div>
    }
}

export default Simulation
