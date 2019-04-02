module.exports = async function(callback) {
  var fs = require('fs');
  var logFile = process.cwd() + `/transactions.json`;

  // remove log file
  fs.unlink(logFile, (err) => {
    if (!err) {
      console.log(logFile, ' was deleted');
    }
  });

  console.log(`creating file ${logFile}`)
  console.log("total number of blocks", await web3.eth.getBlockNumber())

  // start from first block, block 0 is genesis, it doesn't contain any transaction
  for (var i = 1; i <= await web3.eth.getBlockNumber(); i++) {
    var block = await web3.eth.getBlock(i)
    for (var idx = 0; idx < block.transactions.length; idx++)
    var transaction = await web3.eth.getTransaction(block.transactions[idx])
    console.log("from: " + transaction.from + " to:" + transaction.to + " Nonce:" + transaction.nonce)
    var tx_json = JSON.stringify(transaction) + "\n"
    fs.appendFileSync(logFile, tx_json);
  }
}
