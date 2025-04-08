

contract Test {
    function test(address a,bytes calldata data)public view returns (bool b,bytes memory res){
        (b,res)=a.staticcall(data);
    }

    function uinttobytes32(uint a)public view returns (bytes32 b){
        return bytes32(a);

    }
}