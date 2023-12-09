import "./App.css";
import contracts from "./contracts/SecretSpend.json";
import { useEffect, useState } from "react";
import { Contract, ethers, toBigInt } from "ethers";

const contractAddress = "0x9AB81C32e1D621404b253c7fE0fC9972d1645E69";
const contractABI = contracts.abi;

function App() {
    const [currentAccount, setCurrentAccount] = useState(null);
    const [receiverID, setReceiverID] = useState("");
    const [amount, setAmount] = useState("");
    const [showToast, setShowToast] = useState("");

    const checkWalletIsConnected = () => {
        const { ethereum } = window;

        if (!ethereum) {
            console.log("Make sure you have MetaMask!");
            return;
        } else {
            console.log("We have the ethereum object", ethereum);
        }
    };

    const connectWalletHandler = async () => {
        const { ethereum } = window;

        if (!ethereum) {
            alert("Get MetaMask!");
            return;
        }

        try {
            const accounts = await ethereum.request({
                method: "eth_requestAccounts",
            });
            console.log("Connected", accounts[0]);
            setCurrentAccount(accounts[0]);
        } catch (error) {
            console.log(error);
        }
    };

    // const fetchData = async () => {
    //     try {
    //         const response = await fetch("YOUR_API_ENDPOINT");
    //         if (!response.ok) {
    //             throw new Error(`HTTP error! status: ${response.status}`);
    //         }
    //         const data = await response.json();
    //         setApiData(data); // Update state with the API data
    //     } catch (error) {
    //         console.error("Error fetching data:", error);
    //         // Handle the error appropriately
    //     }
    // };

    const proveAndTransferHandler = async (e) => {
        e.preventDefault();

        try {
            setShowToast({ message: "Generating Proof" });
            setTimeout(() => setShowToast(""), 3000);

            const fromIndex = 0;
            const toIndex = receiverID;
            const transferAmount = amount;
            const url = `http://localhost:8080/transfer-funds?fromIndex=${fromIndex}&toIndex=${toIndex}&amount=${transferAmount}`;

            const response = await fetch(url, {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                },
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const data = await response.json();
            console.log("Response:", data);

            setShowToast({ message: "Proof Generated", type: "success" });
            setTimeout(() => setShowToast(""), 3000);

            if (window.ethereum && currentAccount) {
                const provider = new ethers.BrowserProvider(window.ethereum);
                const signer = await provider.getSigner();

                const contract = new Contract(
                    contractAddress,
                    contractABI,
                    signer
                );

                console.log("Contract object:", contract);

                const zkProof = {
                    proof: data.proof.map((x) => toBigInt(x)),
                    input: data.inputs.map((x) => toBigInt(x)),
                };

                const tx = await contract.transferPrivately(zkProof);
                await tx.wait();

                console.log("Transaction successful:", tx);

                // Show success toast
                setShowToast({
                    message: "Transfer Successful",
                    type: "success",
                });
                setTimeout(() => setShowToast(""), 3000);
            } else {
                console.log(
                    "Ethereum object not found or no account connected"
                );
            }
        } catch (error) {
            console.error("An error occurred during form submission:", error);
        }
    };

    const connectWalletButton = () => {
        return (
            <button
                onClick={connectWalletHandler}
                className="cta-button connect-wallet-button"
            >
                Connect Wallet
            </button>
        );
    };

    function determineToastClass(toast) {
        if (toast.type === "success") return "bg-success text-white";
        if (toast.type === "error") return "bg-danger text-white";
        return "bg-primary text-white"; // Default class
    }

    const transferForm = () => {
        const handleReceiverIDChange = (e) => {
            const value = Math.max(1, Math.min(31, Number(e.target.value)));
            setReceiverID(value);
        };

        const handleAmountChange = (e) => {
            setAmount(e.target.value);
        };

        const isFormValid = () => {
            return (
                receiverID !== "" &&
                amount !== "" &&
                !isNaN(receiverID) &&
                !isNaN(amount)
            );
        };
        return (
            <div className="container mt-5">
                {showToast && (
                    <div
                        style={{
                            position: "fixed",
                            bottom: "20px",
                            right: "20px",
                            zIndex: 1050,
                        }}
                    >
                        <div
                            className={`toast show ${determineToastClass(
                                showToast
                            )}`}
                            role="alert"
                            aria-live="assertive"
                            aria-atomic="true"
                        >
                            <div className="toast-header">
                                <strong className="me-auto">
                                    Notification
                                </strong>
                                <button
                                    type="button"
                                    className="btn-close"
                                    onClick={() => setShowToast("")}
                                    aria-label="Close"
                                ></button>
                            </div>
                            <div className="toast-body">
                                {showToast.message}
                            </div>
                        </div>
                    </div>
                )}
                <div className="row justify-content-center">
                    <div className="col-md-4">
                        <form onSubmit={proveAndTransferHandler}>
                            <div className="form-group">
                                <label>ReceiverID:</label>
                                <input
                                    type="number"
                                    value={receiverID}
                                    onChange={handleReceiverIDChange}
                                    className="form-control"
                                    min="0"
                                    max="31"
                                />
                            </div>

                            <div className="form-group">
                                <label>Amount:</label>
                                <input
                                    type="number"
                                    value={amount}
                                    onChange={handleAmountChange}
                                    className="form-control"
                                    min="1"
                                />
                            </div>

                            <button
                                type="submit"
                                className="btn btn-primary btn-block"
                                disabled={!isFormValid()}
                            >
                                Prove & Transfer
                            </button>
                        </form>
                    </div>
                </div>
            </div>
        );
    };

    useEffect(() => {
        checkWalletIsConnected();
    }, []);

    return (
        <div>
            <div className="main-background"></div>
            <div className="container my-5">
                <h1 className="title text-center mb-4">$ecret $pend</h1>
                <div className="row">
                    <div className="col-md-8 offset-md-2">
                        <p className="text-center text-secondary">
                            <strong>$ecret $pend</strong> combines privacy and
                            transactions seamlessly. Utilizing Zero Knowledge
                            Proofs and Additive Homomorphic Encryption, we
                            redefine financial privacy. Your balance remains
                            hidden while you transfer funds securely.
                        </p>
                        <p className="text-center text-secondary">
                            Behind our simple interface is a robust
                            cryptographic system, ensuring each transaction is a
                            secure, private exchange. Experience the pinnacle of
                            financial discretion with $ecret $pend.
                        </p>
                        <div className="text-center mb-4">
                            <a
                                href="https://github.com/shreyas-londhe/secret-spend"
                                target="_blank"
                                rel="noopener noreferrer"
                                className="btn btn-outline-dark"
                            >
                                Check out our GitHub Repository
                            </a>
                        </div>
                    </div>
                </div>
                <div className="d-flex justify-content-center">
                    {currentAccount ? transferForm() : connectWalletButton()}
                </div>
            </div>
        </div>
    );
}

export default App;
