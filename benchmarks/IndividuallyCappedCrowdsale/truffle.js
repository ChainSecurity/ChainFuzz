// module.exports = {
//   // See <http://truffleframework.com/docs/advanced/configuration>
//   // to customize your Truffle configuration!
// };



module.exports = {
    networks: {
      development: {
        host: "127.0.0.1",
        port: 8545,
        network_id: "*" // Match any network id
        // gas: 0xfffffffff,
        // gasPrice: 1
      }
    },
    compilers: {
      solc: {
          version: "0.4.25",    // Fetch exact version from solc-bin (default: truffle's version)
          settings: {          // See the solidity docs for advice about optimization and evmVersion
            optimizer: {
                enabled: true,
                runs: 200
            }
          }
      }
    }
};
