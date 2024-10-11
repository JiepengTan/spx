go install .

rd /s /q ..\temp 
mkdir ..\temp 
xcopy /e /i /y ..\test\* ..\temp 

spx init ../temp
spx run ../temp