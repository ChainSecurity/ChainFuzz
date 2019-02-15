
var SimpleToken = artifacts.require("./SimpleToken.sol");
var IndividuallyCappedCrowdsale = artifacts.require("./IndividuallyCappedCrowdsale.sol");
var Crowdsale = artifacts.require("./Crowdsale.sol");


module.exports = async function(deployer, _, accounts) {

  let token = await SimpleToken.new();
  let individuallyCappedCrowdsale = await IndividuallyCappedCrowdsale.new(new web3.utils.BN(1), accounts[0], token.address);
  await token.transfer(individuallyCappedCrowdsale.address, new web3.utils.BN('1e22'));
  await individuallyCappedCrowdsale.setCap(accounts[0], new web3.utils.BN('1e20'))
  await individuallyCappedCrowdsale.setCap(accounts[1], new web3.utils.BN('1e20'))
};

