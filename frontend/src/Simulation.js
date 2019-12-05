import React, {Component} from 'react';

class SimulationCar extends Component {
    render() {
        var items = [];
        var value;
        for (var key in this.props.car) {
            value = this.props.car[key];
            items.push({"name": key})
        }

        if (this.props.locationState === 0) {
            return <td className={"empty"}>
                <div></div>
            </td>
        }

        if (this.props.locationState === 3) {
            // if(items.length === 0) {
            //     return <td className={"parking"}><div></div></td>
            // }
            return <td className={"parking"}>
                <td>
                    <div>
                        {items.map(item => <div>🚗{item.name}</div>)}
                    </div>
                </td>
            </td>
        }

        if (this.props.locationState === 4) {
            return <td>
                        <div>
                            🔥
                            🔥
                            {items.map(item => <div>🔥🚗{item.name}🔥</div>)}
                            🔥
                            🔥
                        </div>
            </td>
        }
        return (
            <td>
                <div>

                    {items.map(item => <div>🚗{item.name}</div>)}
                </div>
            </td>
        );


        return <td>
            <div>&nbsp;&nbsp;&nbsp;&nbsp;</div>
        </td>
    }
}

class Simulation extends Component {
    render() {
        return <table className={"tableBorder"}>
            {this.props.data.map(row => {
                return <tr>
                    {row.map(item => {
                        return <SimulationCar locationState={item.state} car={item.cars}/>
                    })}
                </tr>
            })}
        </table>
    }
}

export default Simulation
