contract Atest{

    function test(address uniswapV2Pair) public view returns(bytes32 slot0){
        //unchecked{payable(msg.sender).transfer(address(this).balance);}
        assembly{
            slot0 := sload(add(uniswapV2Pair, 0))
        }
    }
}

